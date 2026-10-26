'use strict';
const { test } = require('node:test');
const assert = require('node:assert/strict');
const { parseTag, compareTags, bump, branchForIssue, validatePullRequest, planVersion, publishVersion } = require('../.github/scripts/gitflow.cjs');
const issue = { number: 12, state: 'open', title: 'Definir conexión y validación', labels: [{ name: 'type:feature' }] };
const pull = () => ({ number: 25, title: 'feat(api): definir conexión (#12)', body: '## Cambio\nDescripción\n\nCloses #12\n', base: { ref: 'develop' }, head: { ref: 'feature/12-definir-conexion-y-validacion' }, labels: [{ name: 'version:patch' }] });

test('versiones: incremento SemVer real y prefijo V exacto', () => {
  assert.deepEqual(parseTag('V0.0.1'), [0, 0, 1]);
  assert.equal(bump('V0.0.9'), 'V0.0.10');
  assert.equal(bump('V0.0.9', 'minor'), 'V0.1.0');
  assert.equal(bump('V0.1.9', 'major'), 'V1.0.0');
  assert.ok(compareTags('V0.0.10', 'V0.0.9') > 0);
  for (const tag of ['v0.0.1', 'V01.0.1', 'V0.0.1-dev', 'V0.0']) assert.throws(() => parseTag(tag));
});
test('rama del issue: número, acentos y entradas sin ejecutar', () => {
  assert.equal(branchForIssue(issue), 'feature/12-definir-conexion-y-validacion');
  assert.equal(branchForIssue({ ...issue, title: 'feat(api): Definición $(touch /tmp/evil)' }), 'feature/12-definicion-touch-tmp-evil');
  assert.equal(branchForIssue({ ...issue, title: '🔥' }), 'feature/12-tarea');
  assert.throws(() => branchForIssue({ ...issue, state: 'closed' }));
  assert.throws(() => branchForIssue({ ...issue, labels: ['type:feature', 'type:fix'] }));
});
test('PR válido referencia issue y exige un incremento de versión', () => {
  assert.deepEqual(validatePullRequest(pull(), issue), { number: 12, level: 'patch' });
});
test('PR inválido: main, número, issue cerrado, tipo, título, vínculo y etiquetas', () => {
  const mutations = [
    pr => { pr.base.ref = 'main'; },
    pr => { pr.head.ref = 'feature/sin-numero'; },
    pr => { pr.head.ref = 'feature/13-otro'; },
    pr => { pr.head.ref = 'fix/12-otro'; },
    pr => { pr.title = 'cambios varios'; },
    pr => { pr.title = 'feat(api)!: incompatibilidad (#12)'; },
    pr => { pr.body = 'Closes #123'; },
    pr => { pr.labels = []; },
    pr => { pr.labels.push({ name: 'version:minor' }); },
  ];
  for (const mutate of mutations) { const pr = pull(); mutate(pr); assert.throws(() => validatePullRequest(pr, issue)); }
  assert.throws(() => validatePullRequest(pull(), { ...issue, state: 'closed' }));
  assert.doesNotThrow(() => validatePullRequest(pull(), { ...issue, state: 'closed' }, true));
});
function mockGitHub({ tagged = 'baseline', merged = true } = {}) {
  return {
    paginate: async (method, params) => method(params),
    rest: {
      git: {
        listMatchingRefs: async () => [{ ref: 'refs/tags/V0.0.1', object: { type: 'commit', sha: tagged } }],
        getRef: async () => ({ data: { object: { sha: 'current' } } }),
      },
      pulls: { get: async () => ({ data: { ...pull(), merged: true, merge_commit_sha: 'current' } }) },
      issues: { get: async () => ({ data: issue }) },
      repos: {
        listPullRequestsAssociatedWithCommit: async () => merged ? [{ ...pull(), merged_at: 'date', merge_commit_sha: 'current' }] : [],
        compareCommits: async () => ({ data: { status: 'ahead' } }),
      },
    },
  };
}
const context = { eventName: 'push', ref: 'refs/heads/develop', sha: 'current', repo: { owner: 'Luisnativii', repo: 'ServicioHorasSociales' }, payload: {} };
test('baseline ya etiquetada no aumenta versión', async () => {
  const plan = await planVersion({ github: mockGitHub({ tagged: 'current' }), context });
  assert.equal(plan.tag, 'V0.0.1'); assert.equal(plan.release, false);
});
test('preview de PR calcula versión y no publica tags', async () => {
  const plan = await planVersion({ github: mockGitHub(), context: { ...context, eventName: 'pull_request', payload: { pull_request: { number: 25 } } } });
  assert.equal(plan.tag, 'V0.0.2'); assert.equal(plan.release, false);
});
test('push de PR integrado propone V0.0.2; push directo se rechaza', async () => {
  const plan = await planVersion({ github: mockGitHub(), context });
  assert.equal(plan.tag, 'V0.0.2'); assert.equal(plan.release, true);
  await assert.rejects(planVersion({ github: mockGitHub({ merged: false }), context }));
});
test('ejecución obsoleta no etiqueta y PR nunca obtiene publicación', async () => {
  const github = mockGitHub(); github.rest.git.getRef = async () => ({ data: { object: { sha: 'newer' } } });
  const core = { info: () => {} };
  assert.deepEqual(await publishVersion({ github, context, core, plan: { release: true } }), { skipped: true });
  await assert.rejects(publishVersion({ github, context: { ...context, eventName: 'pull_request' }, core, plan: { release: true } }));
});
test('publicación crea tag anotado una sola vez, sin commits de versión', async () => {
  const github = mockGitHub(); const writes = [];
  github.rest.git.createTag = async args => { writes.push(args); return { data: { sha: 'annotated' } }; };
  github.rest.git.createRef = async args => { writes.push(args); };
  const core = { summary: { addHeading() { return this; }, addRaw() { return this; }, async write() {} } };
  await publishVersion({ github, context, core, plan: { release: true, previous: 'V0.0.1', tag: 'V0.0.2' } });
  assert.equal(writes[0].object, 'current'); assert.equal(writes[1].ref, 'refs/tags/V0.0.2');
});

test('tag existente: reintento idempotente y rechazo de otro commit', async () => {
  const core = { info() {} };
  const plan = { release: true, previous: 'V0.0.0', tag: 'V0.0.1' };
  assert.deepEqual(await publishVersion({ github: mockGitHub({ tagged: 'current' }), context, core, plan }), { tag: 'V0.0.1', existing: true });
  await assert.rejects(publishVersion({ github: mockGitHub(), context, core, plan }), /no se sobrescribirá/);
});

test('issue listo: creación controlada, reutilización y permiso de escritura', async () => {
  const { manageIssue } = require('../.github/scripts/gitflow.cjs');
  for (const reuse of [false, true]) {
    const writes = [];
    const github = {
      paginate: async (method, args) => method(args),
      rest: {
        repos: { getCollaboratorPermissionLevel: async () => ({ data: { permission: 'write' } }) },
        git: {
          listMatchingRefs: async () => reuse ? [{ ref: 'refs/heads/feature/12-ya-existe' }] : [],
          getRef: async () => ({ data: { object: { sha: 'develop-head' } } }),
          createRef: async args => { writes.push(['branch', args]); },
        },
        issues: {
          get: async () => ({ data: issue }),
          removeLabel: async () => {}, addLabels: async () => {}, listComments: async () => [],
          createComment: async args => { writes.push(['comment', args]); },
        },
      },
    };
    const event = { ...context, actor: 'Luisnativii', payload: { action: 'labeled', label: { name: 'status:ready' }, issue: { number: 12 } } };
    await manageIssue({ github, context: event, core: { info() {} } });
    assert.equal(writes.filter(([type]) => type === 'branch').length, reuse ? 0 : 1);
    const comment = writes.find(([type]) => type === 'comment')[1].body;
    assert.ok(comment.includes(reuse ? 'feature/12-ya-existe' : 'feature/12-definir-conexion-y-validacion'));
    github.rest.repos.getCollaboratorPermissionLevel = async () => ({ data: { permission: 'read' } });
    await assert.rejects(manageIssue({ github, context: event, core: { info() {} } }), /escritura/);
  }
});
