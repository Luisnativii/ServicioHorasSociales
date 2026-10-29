import { useEffect, useRef, useState, type FormEvent } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useSession } from '../../app/session';
import { ApiError, NetworkError } from '../../lib/api';
import { authorizedDestination } from '../../app/navigation';

export function accessError(error: unknown): string {
  if (error instanceof NetworkError) return 'No pudimos conectar con el portal. Revisa tu conexión y vuelve a intentarlo.';
  if (error instanceof ApiError) {
    if (error.status === 401) return 'No pudimos iniciar sesión. Revisa tu correo y contraseña.';
    if (error.status === 429) return error.retryAfter
      ? `Hay demasiados intentos. Vuelve a intentarlo en aproximadamente ${Math.ceil(error.retryAfter)} segundos.`
      : 'Hay demasiados intentos. Espera unos minutos antes de volver a intentarlo.';
    if (error.status === 403) return 'No se pudo validar el acceso. Vuelve a intentarlo desde esta página.';
  }
  return 'El portal no pudo completar el acceso. Inténtalo de nuevo más tarde.';
}

export default function Login() {
  const session = useSession();
  const location = useLocation();
  const [correo, setCorreo] = useState('');
  const [password, setPassword] = useState('');
  const [visible, setVisible] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const sending = useRef(false);
  const errorBox = useRef<HTMLParagraphElement>(null);
  useEffect(() => { document.title = 'Iniciar sesión · Horas Sociales'; }, [location.key]);
  useEffect(() => { if (error) errorBox.current?.focus(); }, [error]);
  useEffect(() => {
    const hide = () => { setPassword(''); setVisible(false); };
    window.addEventListener('pagehide', hide);
    return () => window.removeEventListener('pagehide', hide);
  }, []);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (sending.current) return;
    sending.current = true; setSubmitting(true); setError('');
    try { await session.login(correo, password); }
    catch (failure) { setError(accessError(failure)); }
    finally { setPassword(''); setVisible(false); sending.current = false; setSubmitting(false); }
  }
  if (session.status === 'authenticated') return <Navigate to={authorizedDestination(location.state?.from)} replace />;
  const checking = session.status === 'loading' && !submitting;
  return <main className="archivo-access">
    <section className="access-intro" aria-label="Portal de Horas Sociales">
      <a className="access-brand" href="/login" aria-label="Horas Sociales, inicio">
        <span className="brand-mark" aria-hidden="true">hs<span>.</span></span>
        <span>Portal de<br />Horas Sociales</span>
      </a>
      <div className="access-statement">
        <p className="access-eyebrow">Servicio social</p>
        <p className="access-heading">Cada proyecto.<br />Cada persona.<br /><em>Cada hora.</em></p>
        <p className="access-description">Un lugar para organizar el trabajo y conservar su historia.</p>
      </div>
      <p className="access-footnote">Gestión de horas sociales <span aria-hidden="true">/</span> Acceso administrativo</p>
    </section>
    <section className="access-form-area" aria-labelledby="access-title">
      <div className="access-form-wrap">
        <p className="access-eyebrow">Bienvenido de nuevo</p>
        <h1 id="access-title">Iniciar sesión</h1>
        <p className="access-description">Ingresa con tu cuenta de administrador.</p>
        {location.state?.logoutUncertain === true && <p role="alert" className="access-notice">No se pudo confirmar el cierre en el servidor. Comprueba tu conexión antes de continuar.</p>}
        <form onSubmit={event => void submit(event)} aria-busy={submitting || checking} className="access-form">
          <div>
            <label htmlFor="correo">Correo electrónico</label>
            <input id="correo" name="correo" type="email" required autoComplete="username" autoCapitalize="none" spellCheck={false}
              value={correo} onChange={event => setCorreo(event.target.value)} disabled={submitting || checking} />
          </div>
          <div>
            <label htmlFor="contrasena">Contraseña</label>
            <div className="password-field">
              <input id="contrasena" name="contrasena" type={visible ? 'text' : 'password'} required autoComplete="current-password"
                value={password} onChange={event => setPassword(event.target.value)} disabled={submitting || checking} />
              <button type="button" className="password-toggle" aria-label={visible ? 'Ocultar contraseña' : 'Mostrar contraseña'}
                aria-pressed={visible} onClick={() => setVisible(!visible)} disabled={submitting || checking}>{visible ? 'Ocultar' : 'Mostrar'}</button>
            </div>
          </div>
          {error && <p ref={errorBox} tabIndex={-1} role="alert" className="access-error">{error}</p>}
          {session.status === 'error' && !error && <p role="status" className="access-notice">No pudimos comprobar tu sesión anterior. Puedes volver a iniciar sesión.</p>}
          <button className="access-primary" type="submit" disabled={submitting || checking}>
            {submitting ? 'Ingresando…' : checking ? 'Comprobando sesión…' : 'Ingresar'}
            {!submitting && !checking && <span aria-hidden="true">→</span>}
          </button>
        </form>
        <p className="access-help">Acceso exclusivo para personal administrador.</p>
      </div>
      <p className="access-version">Horas Sociales <span aria-hidden="true">·</span> {import.meta.env.VITE_APP_VERSION || '0.0.1'}-dev</p>
    </section>
  </main>;
}
