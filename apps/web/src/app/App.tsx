import { Route, Routes } from 'react-router-dom';
import Login from '../features/auth/Login';
import ServiceStatus from './ServiceStatus';
import AdminLayout, { Home, MissingPage, SessionGuard } from './AdminLayout';

export default function App() {
  return <Routes>
    <Route path="/login" element={<Login />} />
    <Route path="/estado" element={<main className="public-status"><ServiceStatus /></main>} />
    <Route element={<SessionGuard />}>
      <Route element={<AdminLayout />}>
        <Route index element={<Home />} />
        <Route path="/sistema/estado" element={<ServiceStatus />} />
        <Route path="*" element={<MissingPage />} />
      </Route>
    </Route>
  </Routes>;
}
