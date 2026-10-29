import { useEffect, useRef, useState } from 'react';
import { health } from '../lib/health';

type ConnectionState = { api: boolean; db: boolean };

export default function App() {
  const [state, setState] = useState<ConnectionState | null>(null);
  const [loading, setLoading] = useState(true);
  const active = useRef<AbortController | null>(null);

  async function check() {
    active.current?.abort();
    const controller = new AbortController();
    active.current = controller;
    setLoading(true);
    const timer = window.setTimeout(() => controller.abort(), 5000);
    try {
      const [api, db] = await Promise.all([
        health('live', controller.signal),
        health('ready', controller.signal),
      ]);
      if (active.current === controller) setState({ api, db });
    } finally {
      window.clearTimeout(timer);
      if (active.current === controller) setLoading(false);
    }
  }

  useEffect(() => {
    void check();
    return () => {
      active.current?.abort();
      active.current = null;
    };
  }, []);

  const services = [
    { name: 'Backend', available: state?.api, description: 'API del portal' },
    { name: 'PostgreSQL', available: state?.db, description: 'Conexión a la base de datos' },
  ];

  return (
    <main className="min-h-screen bg-slate-950 px-6 py-16 text-slate-100 sm:px-16 sm:py-24">
      <div className="mx-auto max-w-3xl">
        <p className="text-sm font-medium uppercase tracking-widest text-teal-300">Entorno local · {import.meta.env.VITE_APP_VERSION || '0.0.1'}-dev</p>
        <h1 className="mb-5 mt-5 text-4xl font-semibold tracking-tight sm:text-5xl">Portal de Horas Sociales</h1>
        <p className="max-w-xl text-lg leading-relaxed text-slate-300">Un punto de partida para organizar el servicio social. Comprueba aquí la conexión de los servicios del portal.</p>
        <section aria-label="Estado de los servicios" aria-live="polite" aria-busy={loading} className="my-10 grid gap-4 sm:grid-cols-2">
          {services.map(({ name, available, description }) => (
            <article key={name} aria-label={name} className="rounded-2xl border border-slate-700 bg-slate-900 p-6">
              <h2 className="mb-2 text-xl font-semibold">{name}</h2>
              <p className="mb-6 text-sm text-slate-400">{description}</p>
              <p className={`flex items-center gap-2 font-medium ${loading ? 'text-slate-300' : available ? 'text-teal-300' : 'text-amber-300'}`}>
                <span aria-hidden="true" className={`h-2 w-2 rounded-full bg-current ${loading ? 'animate-pulse' : ''}`} />
                {loading ? 'Comprobando…' : state === null ? 'Sin comprobar' : available ? 'Disponible' : 'No disponible'}
              </p>
            </article>
          ))}
        </section>
        {!loading && state && (!state.api || !state.db) && (
          <p role="alert" className="mb-6 rounded-xl border border-amber-800 bg-amber-950/40 p-4 text-amber-200">
            {state.api ? 'La API responde, pero no puede conectar con PostgreSQL.' : 'No se pudo contactar con el backend del portal.'} Vuelve a comprobar cuando el servicio se recupere.
          </p>
        )}
        <button type="button" disabled={loading} onClick={() => void check()} className="rounded-lg bg-teal-300 px-5 py-3 font-semibold text-slate-950 transition hover:bg-teal-200 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-teal-300 disabled:cursor-wait disabled:opacity-50">
          {loading ? 'Comprobando…' : 'Volver a comprobar'}
        </button>
        <p className="mt-5 text-sm text-slate-500">Los estados se consultan al abrir esta página y al pulsar el botón.</p>
      </div>
    </main>
  );
}
