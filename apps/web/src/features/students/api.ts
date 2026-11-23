import { api, ApiError, NetworkError } from '../../lib/api';
import { listCatalog, type Career, type Faculty } from '../academics/api';

type Reference = { id: string; nombre: string };
export type Student = {
  id: string;
  carnet: string;
  nombre_completo: string;
  nombres: string | null;
  apellidos: string | null;
  correo_institucional: string;
  carrera: Reference;
  facultad: Reference;
  activo: boolean;
};
type Period = {
  carrera: Reference;
  facultad: Reference;
  desde: string | null;
  hasta: string | null;
  motivo: string | null;
};
export type StudentDetail = Student & { vigencias: Period[]; historial_disponible: boolean };
export type StudentInput = {
  carnet: string;
  nombres: string | null;
  apellidos: string | null;
  correo_institucional: string;
  carrera_id: string;
  motivo_cambio_carrera: string;
};
type StudentPage = {
  estudiantes: Student[];
  pagina: { page: number; page_size: number; total: number };
};

const base = '/api/v1/estudiantes';
export const studentKey = ['students'];

export function listStudents(filters: URLSearchParams, signal: AbortSignal) {
  return api.request<StudentPage>(`${base}?${filters}`, { signal });
}
export function getStudent(id: string, signal: AbortSignal) {
  return api.request<StudentDetail>(`${base}/${encodeURIComponent(id)}`, { signal });
}
export function saveStudent(input: StudentInput, id?: string) {
  return api.request<Student>(id ? `${base}/${encodeURIComponent(id)}` : base, {
    method: id ? 'PUT' : 'POST', body: input,
  });
}

export async function academicOptions(signal: AbortSignal) {
  async function all(kind: 'facultades' | 'carreras') {
    const entries = [];
    for (let page = 1; ; page++) {
      const params = new URLSearchParams({ page: String(page), page_size: '100', incluir_inactivas: 'true' });
      const result = await listCatalog(kind, params, signal);
      entries.push(...result.entries);
      if (!result.entries.length || page * 100 >= result.pagination.total) return entries;
    }
  }
  const [faculties, careers] = await Promise.all([all('facultades'), all('carreras')]);
  return { faculties: faculties as Faculty[], careers: careers as Career[] };
}

export function studentError(error: Error) {
  if (error instanceof NetworkError) return 'Se perdió la conexión. Comprueba el expediente antes de volver a guardar.';
  if (error instanceof ApiError) {
    if (error.status === 404) return 'El expediente no está disponible.';
    if (error.status === 409) return error.field === 'correo_institucional'
      ? 'Ya existe un estudiante con ese correo institucional.' : 'Ya existe un estudiante con ese carnet.';
    if (error.status === 503 && error.field === 'historial_academico') return 'El core aún no tiene habilitado el historial académico para esta operación.';
    if (error.status === 400) return 'Revisa los datos del formulario y el motivo del cambio de carrera.';
  }
  return 'No se pudo completar la solicitud.';
}
