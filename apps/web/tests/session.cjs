const { test } = require('node:test');
const assert = require('node:assert/strict');
const { ApiClient, ApiError, NetworkError, StaleSessionError } = require('/tmp/hs-web-session-tests/lib/api.js');
const { SessionController, createQueryClient } = require('/tmp/hs-web-session-tests/app/session.js');
const { authorizedDestination, navigation } = require('/tmp/hs-web-session-tests/app/navigation.js');
const json = (status, value) => new Response(JSON.stringify(value), { status });
const error = status => json(status, { error: { code: 'controlled' } });
const deferred = () => { let resolve; const promise = new Promise(r => { resolve = r; }); return { promise, resolve }; };

test('parallel 401 clears cache once and never retries or refreshes', async () => {
  const gate = deferred(); let calls = 0, expired = 0;
  const client = new ApiClient(async () => { calls++; await gate.promise; return error(401); }, () => 'csrf');
  const queries = createQueryClient(); const session = new SessionController(queries, client);
  queries.setQueryData(['private'], { secret: 'fixture' });
  client.onUnauthorized(() => expired++);
  const requests = [client.request('/api/v1/a'), client.request('/api/v1/b')];
  gate.resolve();
  await Promise.allSettled(requests);
  assert.equal(calls, 2); assert.equal(expired, 1);
  assert.equal(session.getSnapshot().status, 'anonymous'); assert.equal(queries.getQueryCache().getAll().length, 0);
});

test('network/403/429 are distinct and writes are sent once', async () => {
  for (const status of [403, 429]) {
    let calls = 0;
    const client = new ApiClient(async (_path, init) => {
      calls++; assert.equal(init.credentials, 'same-origin'); assert.equal(init.cache, 'no-store');
      assert.equal(init.headers.get('X-CSRF-Token'), 'proof');
      return new Response('{}', { status, headers: { 'Retry-After': '60' } });
    }, () => 'proof');
    await assert.rejects(client.request('/api/v1/write', { method: 'POST', body: { value: 1 } }), e => e instanceof ApiError && e.status === status && (status !== 429 || e.retryAfter === 60));
    assert.equal(calls, 1);
  }
  const client = new ApiClient(async () => { throw new Error('network fixture'); }, () => 'proof');
  await assert.rejects(client.request('/api/v1/write', { method: 'POST' }), NetworkError);
});

test('CSRF preparation is shared; header reads rotated cookie', async () => {
  let cookie, csrfCalls = 0; const gate = deferred(); const proofs = [];
  const client = new ApiClient(async (path, init) => {
    if (path.endsWith('/csrf')) { csrfCalls++; await gate.promise; cookie = 'prepared'; return new Response(null, { status: 204 }); }
    proofs.push(init.headers.get('X-CSRF-Token')); return new Response(null, { status: 204 });
  }, () => cookie);
  const one = client.request('/api/v1/a', { method: 'POST' });
  const two = client.request('/api/v1/b', { method: 'POST' });
  gate.resolve(); await Promise.all([one, two]); assert.equal(csrfCalls, 1);
  cookie = 'rotated'; await client.request('/api/v1/c', { method: 'POST' });
  assert.deepEqual(proofs, ['prepared', 'prepared', 'rotated']);
});

test('late data from a previous identity is discarded', async () => {
  const gate = deferred(); const client = new ApiClient(() => gate.promise, () => 'proof');
  const result = client.request('/api/v1/private'); client.invalidate(); gate.resolve(json(200, { private: 'fixture' }));
  await assert.rejects(result, StaleSessionError);
});

test('public health is unaffected by an anonymous session check', async () => {
  const gate = deferred(); const client = new ApiClient(() => gate.promise, () => 'proof');
  const result = client.request('/api/health/live', { authenticated: false });
  client.invalidate(); gate.resolve(json(200, { status: 'ok' }));
  assert.deepEqual(await result, { status: 'ok' });
});

test('login and logout prepare CSRF, rotate identity and clear private caches', async () => {
  let cookie, identity = { id: 'fixture-id', correo: 'fixture@example.test' }; const paths = [];
  const client = new ApiClient(async (path, init) => {
    paths.push(path);
    if (path.endsWith('/csrf')) { cookie = 'prepared'; return new Response(null, { status: 204 }); }
    if (path.endsWith('/login')) { assert.equal(init.headers.get('X-CSRF-Token'), 'prepared'); cookie = 'bound'; return json(200, identity); }
    if (path.endsWith('/logout')) { cookie = undefined; return new Response(null, { status: 204 }); }
    return json(200, identity);
  }, () => cookie);
  const queries = createQueryClient(); const session = new SessionController(queries, client);
  await session.login('fixture@example.test', 'synthetic1');
  assert.deepEqual(session.getSnapshot().user, identity);
  queries.setQueryData(['private'], 'fixture');
  await session.logout(); assert.equal(session.getSnapshot().status, 'anonymous');
  assert.equal(queries.getQueryCache().getAll().length, 0);
  assert.deepEqual(paths, ['/api/v1/auth/csrf', '/api/v1/auth/login', '/api/v1/auth/me', '/api/v1/auth/csrf', '/api/v1/auth/logout']);
});

test('session validation deduplicates requests and keeps network failure distinct from anonymous', async () => {
  const gate = deferred(); let calls = 0;
  const client = new ApiClient(async () => { calls++; await gate.promise; throw new Error('offline'); }, () => 'proof');
  const queries = createQueryClient(); const session = new SessionController(queries, client);
  const one = session.validate(), two = session.validate(); assert.equal(one, two);
  gate.resolve(); await one; assert.equal(calls, 1); assert.equal(session.getSnapshot().status, 'error');
  assert.ok(session.getSnapshot().error instanceof NetworkError);
});

test('CSRF requests serialize across session changes', async () => {
  const gate = deferred(); let cookie, calls = 0;
  const client = new ApiClient(async () => {
    calls++; if (calls === 1) await gate.promise;
    cookie = calls === 1 ? 'old' : 'current'; return new Response(null, { status: 204 });
  }, () => cookie);
  const old = client.prepareCSRF(); client.invalidate(); const current = client.prepareCSRF();
  const rejected = assert.rejects(old, StaleSessionError);
  gate.resolve(); await rejected; await current;
  assert.equal(calls, 2); assert.equal(cookie, 'current');
});

test('uncertain logout clears local data without retrying the mutation', async () => {
  let cookie, writes = 0;
  const client = new ApiClient(async path => {
    if (path.endsWith('/csrf')) { cookie = 'prepared'; return new Response(null, { status: 204 }); }
    writes++; throw new Error('connection lost');
  }, () => cookie);
  const queries = createQueryClient(); const session = new SessionController(queries, client);
  queries.setQueryData(['private'], 'fixture');
  await assert.rejects(session.logout(), NetworkError);
  assert.equal(writes, 1); assert.equal(session.getSnapshot().status, 'anonymous');
  assert.equal(queries.getQueryCache().getAll().length, 0);
});

test('login confirms me for two distinct accounts before exposing identity', async () => {
  for (const id of ['first', 'second']) {
    const gate = deferred();
    const client = new ApiClient(async path => {
      if (path.endsWith('/csrf')) return new Response(null, { status: 204 });
      if (path.endsWith('/login')) return json(200, { id, correo: `${id}@example.test` });
      return gate.promise;
    }, () => 'proof');
    const queries = createQueryClient(); const session = new SessionController(queries, client);
    const pending = session.login(`${id}@example.test`, 'synthetic1');
    assert.equal(session.getSnapshot().status, 'loading');
    gate.resolve(json(200, { id, correo: `${id}@example.test` })); await pending;
    assert.equal(session.getSnapshot().user.id, id); queries.clear();
  }
});

test('successful login response without a usable cookie does not authenticate', async () => {
  const client = new ApiClient(async path => {
    if (path.endsWith('/csrf')) return new Response(null, { status: 204 });
    if (path.endsWith('/login')) return json(200, { id: 'fixture', correo: 'fixture@example.test' });
    return error(401);
  }, () => 'proof');
  const queries = createQueryClient(); const session = new SessionController(queries, client);
  await assert.rejects(session.login('fixture@example.test', 'synthetic1'), ApiError);
  assert.equal(session.getSnapshot().status, 'anonymous'); queries.clear();
});

test('return destination is restricted to implemented private routes', () => {
  for (const route of navigation) assert.equal(authorizedDestination(route.path), route.path);
  for (const value of ['https://attacker.test', '//attacker.test', '/login', '/estudiantes', '/sistema/estado?to=elsewhere', undefined]) {
    assert.equal(authorizedDestination(value), '/');
  }
});

test('suspending a cached document hides identity and removes private queries', async () => {
  const client = new ApiClient(async () => json(200, { id: 'fixture', correo: 'fixture@example.test' }), () => 'proof');
  const queries = createQueryClient(); const session = new SessionController(queries, client);
  await session.validate(); queries.setQueryData(['private'], 'fixture');
  session.suspend();
  assert.deepEqual(session.getSnapshot(), { status: 'loading', user: null, error: null });
  assert.equal(queries.getQueryCache().getAll().length, 0);
  await session.validate(); assert.equal(session.getSnapshot().status, 'authenticated'); queries.clear();
});
