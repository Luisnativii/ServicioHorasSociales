import { useEffect, useState } from 'react';
import { Navigate, Route, Routes, useNavigate } from 'react-router-dom';
import { useSession } from './session';
import Login from '../features/auth/Login';
import ServiceStatus from './ServiceStatus';

function PrivateEntry() {
  const session = useSession();
  const navigate = useNavigate();
  const [leaving, setLeaving] = useState(false);
  useEffect(() => { document.title = 'Acceso · Horas Sociales'; }, []);
  if (session.status === 'anonymous') return <Navigate to="/login" replace />;
  if (session.status !== 'authenticated') return <main className="archivo-message"><div>
    <h1>{session.status === 'loading' ? 'Comprobando sesión…' : 'No pudimos comprobar tu sesión'}</h1>
    {session.status === 'error' && <><p>Revisa tu conexión antes de continuar.</p><button className="access-primary" onClick={() => void session.validate()}>Volver a comprobar</button></>}
  </div></main>;
  // Minimal confirmed-identity destination; administrative navigation is #32.
  return <main className="archivo-message"><section>
    <p className="access-eyebrow">Portal de Horas Sociales</p>
    <h1>Sesión iniciada</h1>
    <p>Has ingresado como <strong>{session.user.correo}</strong>.</p>
    <button className="access-primary" disabled={leaving} onClick={async () => {
      setLeaving(true);
      try { await session.logout(); } catch { navigate('/login', { replace: true, state: { logoutUncertain: true } }); }
      finally { setLeaving(false); }
    }}>{leaving ? 'Cerrando sesión…' : 'Cerrar sesión'}</button>
  </section></main>;
}

export default function App() {
  return <Routes>
    <Route path="/login" element={<Login />} />
    <Route path="/estado" element={<ServiceStatus />} />
    <Route path="/" element={<PrivateEntry />} />
    <Route path="*" element={<Navigate to="/" replace />} />
  </Routes>;
}
