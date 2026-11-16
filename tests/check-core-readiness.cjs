const assert = require("node:assert/strict");
const { setTimeout: delay } = require("node:timers/promises");

async function check(service, code, status) {
  const base = `http://${service}:8080/api/health`;
  let live;
  for (let i = 0; i < 15; i++) {
    live = await fetch(`${base}/live`, { signal: AbortSignal.timeout(2000) }).catch(() => null);
    if (live?.ok) break;
    await delay(1000);
  }
  assert.equal(live?.status, 200, `${service} no respondió a liveness`);
  const ready = await fetch(`${base}/ready`, { signal: AbortSignal.timeout(4000) });
  assert.equal(ready.status, code, service);
  assert.deepEqual(await ready.json(), { status });
}

const databaseDown = process.argv.includes("--db-down");
const checks = [
  check("api", databaseDown ? 503 : 200, databaseDown ? "unavailable" : "ok"),
  check("api-missing", 503, "unavailable"),
];
Promise.all(checks).then(() => {
  console.log(databaseDown
    ? "hs-test-core: PostgreSQL detenido, liveness devuelve 200 y readiness 503."
    : "hs-test-core: catálogo accesible devuelve 200 y esquema ausente devuelve 503.");
}).catch(error => {
  console.error(error.message);
  process.exitCode = 1;
});
