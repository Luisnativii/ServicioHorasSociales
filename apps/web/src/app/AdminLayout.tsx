import { useEffect, useRef, useState } from 'react';
import { Link, Navigate, NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useSession } from './session';
import { navigation } from './navigation';

export function SessionGuard() {
  const session = useSession();
  const location = useLocation();
  if (session.status === 'anonymous') return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  if (session.status === 'loading') return <main className="archivo-message"><div role="status"><h1>Comprobando sesión…</h1></div></main>;
  if (session.status === 'error') return <main className="archivo-message"><section>
    <h1>No pudimos comprobar tu sesión</h1><p role="alert">Revisa tu conexión antes de continuar.</p>
    <button className="access-primary" onClick={() => void session.validate()}>Volver a comprobar</button>
    <Link className="shell-text-link" to="/login">Ir al acceso</Link>
  </section></main>;
  return <Outlet />;
}

function Brand() {
  return <Link className="access-brand" to="/" aria-label="Horas Sociales, inicio">
    <span className="brand-mark" aria-hidden="true">hs<span>.</span></span><span>Portal de<br />Horas Sociales</span>
  </Link>;
}

export default function AdminLayout() {
  const session = useSession();
  const location = useLocation();
  const navigate = useNavigate();
  const main = useRef<HTMLElement>(null);
  const menu = useRef<HTMLDialogElement>(null);
  const logoutDialog = useRef<HTMLDialogElement>(null);
  const menuButton = useRef<HTMLButtonElement>(null);
  const logoutButton = useRef<HTMLButtonElement>(null);
  const logoutOpener = useRef<HTMLButtonElement | null>(null);
  const navigatingMenu = useRef(false);
  const sending = useRef(false);
  const [leaving, setLeaving] = useState(false);
  const current = navigation.find(item => item.path === location.pathname);

  useEffect(() => {
    document.title = `${current?.title ?? 'Página no disponible'} · Horas Sociales`;
    main.current?.focus();
  }, [location.pathname, current?.title]);

  function links() {
    return <nav aria-label="Navegación principal" className="shell-nav">
      {navigation.map(item => <NavLink key={item.path} to={item.path} end onClick={() => {
        if (menu.current?.open) { navigatingMenu.current = true; menu.current.close(); }
      }}
        className={({ isActive }) => isActive ? 'shell-nav-link active' : 'shell-nav-link'}>{item.label}</NavLink>)}
    </nav>;
  }

  async function logout() {
    if (sending.current) return;
    sending.current = true; setLeaving(true); logoutDialog.current?.close();
    try { await session.logout(); navigate('/login', { replace: true }); }
    catch { navigate('/login', { replace: true, state: { logoutUncertain: true } }); }
    finally { sending.current = false; setLeaving(false); }
  }

  return <div className="admin-shell">
    <a href="#contenido" className="skip-link">Saltar al contenido</a>
    <aside className="shell-sidebar" aria-label="Menú del portal">
      <Brand />
      <div className="shell-nav-section"><p className="access-eyebrow">Espacio de trabajo</p>{links()}</div>
      <div className="shell-sidebar-bottom">
        <p>Administración</p>
        <button ref={logoutButton} type="button" className="shell-logout" disabled={leaving}
          onClick={() => { logoutOpener.current = logoutButton.current; logoutDialog.current?.showModal(); }}>{leaving ? 'Cerrando sesión…' : 'Cerrar sesión'}</button>
        <small>{import.meta.env.VITE_APP_VERSION || '0.0.1'}-dev</small>
      </div>
    </aside>
    <div className="shell-workspace">
      <header className="shell-header">
        <button ref={menuButton} className="shell-menu-button" aria-haspopup="dialog" aria-controls="menu-movil"
          onClick={() => menu.current?.showModal()}>Menú</button>
        <p className="shell-section-title">{current?.title ?? 'Página no disponible'}</p>
        <div className="shell-account"><span className="shell-account-dot" aria-hidden="true" /><span>{session.user?.correo}</span><small>ADMIN</small></div>
      </header>
      <main id="contenido" ref={main} tabIndex={-1} className="shell-content"><Outlet /></main>
      <footer className="shell-footer">Portal de Horas Sociales</footer>
    </div>

    <dialog id="menu-movil" ref={menu} className="shell-mobile-menu" aria-labelledby="menu-title"
      onClose={() => { if (navigatingMenu.current) { navigatingMenu.current = false; main.current?.focus(); } else menuButton.current?.focus(); }}>
      <div className="shell-dialog-top"><h2 id="menu-title">Navegación</h2><button autoFocus onClick={() => menu.current?.close()}>Cerrar menú</button></div>
      {links()}
      <button className="shell-logout" onClick={() => { menu.current?.close(); logoutOpener.current = menuButton.current; logoutDialog.current?.showModal(); }}>Cerrar sesión</button>
    </dialog>
    <dialog ref={logoutDialog} className="shell-confirm" aria-labelledby="logout-title" aria-describedby="logout-description"
      onClose={() => logoutOpener.current?.focus()}>
      <h2 id="logout-title">¿Cerrar sesión?</h2>
      <p id="logout-description">Para volver a entrar necesitarás tu correo y contraseña.</p>
      <div className="shell-dialog-actions">
        <button className="shell-secondary" autoFocus onClick={() => logoutDialog.current?.close()}>Seguir aquí</button>
        <button className="access-primary" disabled={leaving} onClick={() => void logout()}>Cerrar sesión</button>
      </div>
    </dialog>
  </div>;
}

export function Home() {
  return <section className="shell-home">
    <p className="access-eyebrow">Administración</p>
    <h1>Tu espacio de trabajo</h1>
    <p className="shell-lead">Bienvenido al Portal de Horas Sociales.</p>
    <div className="shell-empty">
      <span className="empty-rule" aria-hidden="true" />
      <h2>El espacio está preparado</h2>
      <p>Los módulos de gestión todavía no están disponibles. Puedes consultar el estado del portal desde el menú.</p>
      <Link className="shell-text-link" to="/sistema/estado">Ver estado del portal <span aria-hidden="true">→</span></Link>
    </div>
  </section>;
}

export function MissingPage() {
  return <section className="shell-home"><p className="access-eyebrow">Página no disponible</p><h1>No encontramos esta página</h1>
    <p className="shell-lead">La dirección no está disponible en el portal.</p><Link className="shell-text-link" to="/">Volver al inicio</Link></section>;
}
