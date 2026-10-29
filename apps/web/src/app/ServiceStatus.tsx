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
    <section className="service-status">
      <p className="access-eyebrow">Disponibilidad</p>
      <h1>Estado del portal</h1>
      <p className="shell-lead">Comprueba la disponibilidad de los servicios.</p>
      <div aria-label="Estado de los servicios" aria-live="polite" aria-busy={loading} className="service-grid">
        {services.map(({ name, available, description }) => (
          <article key={name} aria-label={name}>
            <h2>{name}</h2><p>{description}</p>
            <strong className={available ? 'service-ok' : 'service-wait'}>
              {loading ? 'Comprobando…' : state === null ? 'Sin comprobar' : available ? 'Disponible' : 'No disponible'}
            </strong>
          </article>
        ))}
      </div>
      {!loading && state && (!state.api || !state.db) && <p role="alert" className="access-error">
        {state.api ? 'El portal responde, pero no puede conectar con la base de datos.' : 'No se pudo contactar con el portal.'} Vuelve a comprobar cuando el servicio se recupere.
      </p>}
      <button type="button" disabled={loading} onClick={() => void check()} className="access-primary">
        {loading ? 'Comprobando…' : 'Volver a comprobar'}
      </button>
    </section>
  );
}
