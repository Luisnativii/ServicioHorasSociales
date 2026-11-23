import { api } from '../../lib/api';

type Contact = {
  nombre: string | null;
  correo: string | null;
  telefono: string | null;
  extension: string | null;
};

export type Institution = {
  id: string;
  nombre: string;
  codigo_css: string | null;
  activo: boolean;
  sede: { distrito_id: string; nombre: string } | null;
  ubicacion_heredada: string | null;
  contacto: Contact | null;
};

type InstitutionPage = {
  instituciones: Institution[];
  pagina: { page: number; page_size: number; total: number };
};

export type InstitutionInput = {
  nombre: string;
  distrito_id: string | null;
  contacto?: Contact;
};

const base = '/api/v1/instituciones';
export const institutionKey = ['institutions'];

export function listInstitutions(filters: URLSearchParams, signal: AbortSignal) {
  return api.request<InstitutionPage>(`${base}?${filters}`, { signal });
}

export function getInstitution(id: string, signal: AbortSignal) {
  return api.request<Institution>(`${base}/${encodeURIComponent(id)}`, { signal });
}

export function saveInstitution(input: InstitutionInput, id?: string) {
  return api.request<Institution>(id ? `${base}/${encodeURIComponent(id)}` : base, {
    method: id ? 'PUT' : 'POST',
    body: input,
  });
}

export function deactivateInstitution(id: string) {
  return api.request<Institution>(`${base}/${encodeURIComponent(id)}/desactivar`, { method: 'POST' });
}
