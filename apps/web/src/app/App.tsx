import { Route, Routes } from 'react-router-dom';
import Login from '../features/auth/Login';
import ServiceStatus from './ServiceStatus';
import Accounts from '../features/accounts/Accounts';
import AcademicCatalog from '../features/academics/AcademicCatalog';
import ClassificationCatalog from '../features/classification/ClassificationCatalog';
import TerritoryCatalog from '../features/territory/TerritoryCatalog';
import AdminLayout, { Home, MissingPage, SessionGuard } from './AdminLayout';

export default function App() {
  return <Routes>
    <Route path="/login" element={<Login />} />
    <Route path="/estado" element={<main className="public-status"><ServiceStatus /></main>} />
    <Route element={<SessionGuard />}>
      <Route element={<AdminLayout />}>
        <Route index element={<Home />} />
        <Route path="/administradores" element={<Accounts />} />
        <Route path="/catalogos/academico" element={<AcademicCatalog />} />
        <Route path="/catalogos/territorial" element={<TerritoryCatalog />} />
        <Route path="/catalogos/clasificacion" element={<ClassificationCatalog />} />
        <Route path="/sistema/estado" element={<ServiceStatus />} />
        <Route path="*" element={<MissingPage />} />
      </Route>
    </Route>
  </Routes>;
}
