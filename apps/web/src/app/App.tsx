import { Route, Routes } from 'react-router-dom';
import Login from '../features/auth/Login';
import ServiceStatus from './ServiceStatus';
import Accounts from '../features/accounts/Accounts';
import AcademicCatalog from '../features/academics/AcademicCatalog';
import ClassificationCatalog from '../features/classification/ClassificationCatalog';
import TerritoryCatalog from '../features/territory/TerritoryCatalog';
import Institutions from '../features/institutions/Institutions';
import ActivityReport from '../features/reports/ActivityReport';
import Projects from '../features/projects/Projects';
import Students from '../features/students/Students';
import AdminLayout, { Home, MissingPage, SessionGuard } from './AdminLayout';

export default function App() {
  return <Routes>
    <Route path="/login" element={<Login />} />
    <Route path="/estado" element={<main className="public-status"><ServiceStatus /></main>} />
    <Route element={<SessionGuard />}>
      <Route element={<AdminLayout />}>
        <Route index element={<Home />} />
        <Route path="/administradores" element={<Accounts />} />
        <Route path="/instituciones" element={<Institutions />} />
        <Route path="/reportes" element={<ActivityReport />} />
        <Route path="/proyectos" element={<Projects />} />
        <Route path="/estudiantes" element={<Students />} />
        <Route path="/catalogos/academico" element={<AcademicCatalog />} />
        <Route path="/catalogos/territorial" element={<TerritoryCatalog />} />
        <Route path="/catalogos/clasificacion" element={<ClassificationCatalog />} />
        <Route path="/sistema/estado" element={<ServiceStatus />} />
        <Route path="*" element={<MissingPage />} />
      </Route>
    </Route>
  </Routes>;
}
