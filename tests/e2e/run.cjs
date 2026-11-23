const assert = require('node:assert/strict');
const { chromium, request } = require('playwright');
const { readFileSync } = require('node:fs');
const { execFileSync } = require('node:child_process');
const { performance } = require('node:perf_hooks');

const origin = 'http://127.0.0.1:8080';
const day = new Intl.DateTimeFormat('en-CA', {
  dateStyle: 'short', timeZone: 'America/El_Salvador',
}).format(new Date());
const run = process.env.HS_E2E_RUN || String(Date.now());
let stage = 'arranque';
const button = (page, name) => page.getByRole('button', { name, exact: true });
const heading = (page, name) => page.getByRole('heading', { name, exact: true }).waitFor();

async function login(browser, suffix) {
  stage = 'inicio de sesión de administrador ' + suffix;
  const context = await browser.newContext({ baseURL: origin });
  const page = await context.newPage();
  await page.goto('/login');
  await page.getByLabel('Correo electrónico').fill(`e2e-${suffix}@example.test`);
  await page.getByLabel('Contraseña', { exact: true }).fill(
    readFileSync(`/e2e-secrets/${suffix}.password`, 'utf8'));
  await button(page, 'Ingresar').click();
  await page.waitForURL(origin + '/');
  const me = await context.request.get('/api/v1/auth/me');
  assert.equal(me.status(), 200);
  return { context, page, id: (await me.json()).id };
}

async function selectAny(locator) {
  await locator.locator('option:not([value=""])').first().waitFor({ state: 'attached' });
  await locator.selectOption({ index: 1 });
}

async function journey(browser) {
  stage = 'acceso anónimo';
  const unauthenticated = await request.newContext({ baseURL: origin });
  assert.equal((await unauthenticated.get('/api/v1/reportes/actividad.xlsx')).status(), 401);
  await unauthenticated.dispose();
  const { context, page, id } = await login(browser, 'a');
  const second = await login(browser, 'b');
  assert.notEqual(id, second.id);
  await second.context.close();

  stage = 'protección CSRF';
  const denied = await context.request.post('/api/v1/instituciones', {
    headers: { Origin: origin },
    data: { nombre: 'No guardar E2E80 ' + run },
  });
  assert.equal(denied.status(), 403);

  const org = 'Institución E2E80 ' + run;
  const title = 'Proyecto E2E80 ' + run;
  const carnet = '08' + run.slice(-6);
  stage = 'alta de institución';
  await page.goto('/instituciones');
  await button(page, 'Nueva institución').click();
  await page.getByLabel('Nombre de la institución').fill(org);
  await button(page, 'Guardar institución').click();
  await heading(page, org);

  stage = 'alta de proyecto';
  await page.goto('/proyectos');
  await button(page, 'Nuevo proyecto').click();
  const projectPanel = page.getByRole('heading', { name: 'Nuevo proyecto', exact: true }).locator('..');
  await projectPanel.getByLabel('Título', { exact: true }).fill(title);
  await projectPanel.getByLabel('Descripción', { exact: true }).fill('Datos sintéticos de la prueba integral');
  await projectPanel.locator('select').first().selectOption({ label: org });
  await selectAny(projectPanel.getByLabel('Departamento', { exact: true }));
  await selectAny(projectPanel.getByLabel('Municipio', { exact: true }));
  await selectAny(projectPanel.getByLabel('Distrito', { exact: true }));
  await selectAny(projectPanel.locator('select[name=tipo_id]'));
  await selectAny(projectPanel.locator('select[name=area_id]'));
  await selectAny(projectPanel.locator('select[name=modalidad_id]'));
  await projectPanel.getByLabel('Nombre', { exact: true }).fill('Responsable de prueba integral');
  await button(page, 'Guardar').click();
  await heading(page, title);
  const projectURL = page.url();

  stage = 'alta de expediente';
  await page.goto('/estudiantes');
  await button(page, 'Nuevo expediente').click();
  await page.getByLabel('Carnet', { exact: true }).fill(carnet);
  await page.getByLabel('Nombres', { exact: true }).fill('Prueba');
  await page.getByLabel('Apellidos', { exact: true }).fill('Integral ' + run);
  await page.getByLabel('Correo institucional', { exact: true }).fill(`e2e80-${run}@uca.edu.sv`);
  await selectAny(page.getByLabel('Carrera'));
  await button(page, 'Guardar').click();
  await heading(page, 'Prueba Integral ' + run);
  const studentURL = page.url();

  stage = 'asignación y acreditación';
  await page.goto(projectURL);
  await button(page, 'Asignar estudiante').click();
  await page.getByLabel('Buscar por nombre o carnet', { exact: true }).fill(carnet);
  await button(page, 'Buscar').click();
  await page.getByRole('button', { name: 'Prueba Integral ' + run + ' · ' + carnet, exact: true }).click();
  await button(page, 'Asignar').click();
  await page.getByText('Asignación guardada.', { exact: true }).waitFor();
  await button(page, 'Acreditar horas').click();
  let dialog = page.getByRole('dialog');
  await dialog.getByLabel('Horas', { exact: true }).fill('7');
  await dialog.getByLabel('Concepto', { exact: true }).fill('Movimiento E2E80 ' + run);
  await dialog.getByRole('button', { name: 'Acreditar', exact: true }).click();
  await page.getByText('Horas acreditadas.').waitFor();
  await page.goto(studentURL);
  await page.getByRole('link', { name: title }).first().waitFor();
  await page.getByText('7 de 300 · 293 pendientes', { exact: true }).waitFor();

  stage = 'retiro y cierre';
  await button(page, 'Retirar estudiante').click();
  dialog = page.getByRole('dialog');
  await dialog.getByLabel('Motivo', { exact: true }).fill('Retiro E2E80 ' + run);
  await dialog.getByRole('button', { name: 'Confirmar cierre' }).click();
  await page.getByText('Retirada', { exact: true }).waitFor();
  await page.goto(projectURL);
  await page.getByText('ABANDONADO', { exact: true }).waitFor();
  assert.equal(await button(page, 'Asignar estudiante').count(), 0);

  stage = 'dashboard y descarga';
  await page.goto('/reportes');
  await page.getByLabel('Actividad desde', { exact: true }).fill(day);
  await page.getByLabel('Actividad hasta', { exact: true }).fill(day);
  await page.locator('select[name=estado]').selectOption('ABANDONADO');
  await button(page, 'Aplicar filtros').click();
  await page.getByRole('row').filter({ hasText: title }).waitFor();
  const query = new URLSearchParams({ desde: day, hasta: day, estado: 'ABANDONADO' });
  const dashboard = await context.request.get('/api/v1/reportes/actividad?' + query);
  assert.equal(dashboard.status(), 200);
  const { resumen } = await dashboard.json();
  const downloaded = page.waitForEvent('download');
  await button(page, 'Descargar Excel').click();
  const download = await downloaded;
  const archive = await download.path();
  const summary = execFileSync('unzip', ['-p', archive, 'xl/worksheets/sheet1.xml'], { encoding: 'utf8' });
  const movements = execFileSync('unzip', ['-p', archive, 'xl/worksheets/sheet2.xml'], { encoding: 'utf8' });
  assert.ok(movements.includes(carnet));
  assert.ok(movements.includes(title));
  assert.ok(summary.includes(`<c r="B6" t="n"><v>${resumen.externas}</v></c>`));

  stage = 'cierre de sesión';
  await button(page, 'Cerrar sesión').first().click();
  const logoutResponse = page.waitForResponse(response => response.url() === origin + '/api/v1/auth/logout');
  await page.getByRole('dialog').getByRole('button', { name: 'Cerrar sesión', exact: true }).click();
  assert.equal((await logoutResponse).status(), 204);
  await page.waitForURL(/\/login/);
  assert.equal((await context.request.get('/api/v1/auth/me')).status(), 401);
  await context.close();
  console.log('Recorrido Chromium aprobado: dos ADMIN, CSRF, instituciones, proyectos, expedientes, asignación, retiro y cierre, horas, reportes, XLSX y logout.');
}

async function benchmark(browser) {
  const { context } = await login(browser, 'a');
  const storageState = await context.storageState();
  const clients = await Promise.all(Array.from({ length: 5 }, () =>
    request.newContext({ baseURL: origin, storageState })));
  const activity = '/api/v1/reportes/actividad';
  const filters = 'desde=2024-02-01&hasta=2024-02-01';
  const scenarios = [
    ['estudiantes_25', '/api/v1/estudiantes?q=Bench80&page_size=25'],
    ['proyectos_25', '/api/v1/proyectos?q=Bench80&page_size=25'],
    ['dashboard_5000', activity + '?' + filters],
    ['xlsx_5000', activity + '.xlsx?' + filters],
  ];
  try {
    for (const [name, path] of scenarios) {
      stage = 'medir ' + name;
      const samples = [];
      let failures = 0;
      let bytes = 0;
      async function batch(record) {
        await Promise.all(clients.map(async client => {
          const start = performance.now();
          try {
            const response = await client.get(path, { timeout: 35000 });
            const payload = await response.body();
            if (record) {
              if (!response.ok()) failures++;
              else {
                samples.push(performance.now() - start);
                bytes = Math.max(bytes, payload.length);
              }
            }
            if (!record) assert.equal(response.status(), 200);
            await response.dispose();
          } catch (error) {
            if (record) failures++;
            else throw error;
          }
        }));
      }
      await batch(false);
      for (let batchNumber = 0; batchNumber < 20; batchNumber++) await batch(true);
      samples.sort((a, b) => a - b);
      function p(q) {
        return samples.length ? Math.round(samples[Math.ceil(samples.length * q) - 1]) : null;
      }
      console.log(JSON.stringify({ escenario: name, muestras: samples.length, errores: failures,
        p50_ms: p(0.5), p95_ms: p(0.95), max_ms: p(1), bytes_max: bytes }));
      assert.equal(failures, 0);
    }
  } finally {
    await Promise.all(clients.map(client => client.dispose()));
    await context.close();
  }
}

(async () => {
  const browser = await chromium.launch();
  try {
    if (process.argv[2] === 'benchmark') await benchmark(browser);
    else await journey(browser);
  } finally { await browser.close(); }
})().catch(error => {
  // Do not dump Playwright call logs, cookies or input values.
  const location = error.stack?.match(/run\.cjs:\d+:\d+/)?.[0];
  console.error(`Falló ${stage} (${error.name}, ${location ?? 'sin ubicación'}).`);
  process.exitCode = 1;
});
