import { api } from '../../lib/api';

export type Faculty = { id: string; nombre: string; activo: boolean };
export type Career = Faculty & { facultad_id: string; facultad_activa: boolean };
export type CatalogKind = 'facultades' | 'carreras';
export type CatalogEntry = Faculty | Career;
type Pagination = { page: number; page_size: number; total: number };

const base = '/api/v1/catalogos';

export async function listCatalog(kind: CatalogKind, params: URLSearchParams, signal: AbortSignal) {
  if (kind === 'facultades') {
    const result = await api.request<{ facultades: Faculty[]; pagina: Pagination }>(`${base}/facultades?${params}`, { signal });
    return { entries: result.facultades, pagination: result.pagina };
  }
  const result = await api.request<{ carreras: Career[]; pagina: Pagination }>(`${base}/carreras?${params}`, { signal });
  return { entries: result.carreras, pagination: result.pagina };
}

export function getCatalogEntry(kind: CatalogKind, id: string, signal: AbortSignal) {
  return api.request<CatalogEntry>(`${base}/${kind}/${encodeURIComponent(id)}`, { signal });
}

export type CareerCoordinator = { carrera_id: string; coordinador_nombre: string | null };

export function getCareerCoordinator(id: string, signal: AbortSignal) {
  return api.request<CareerCoordinator>(`${base}/carreras/${encodeURIComponent(id)}/coordinador`, { signal });
}

export function saveCareerCoordinator(id: string, name: string | null) {
  return api.request<CareerCoordinator>(`${base}/carreras/${encodeURIComponent(id)}/coordinador`, {
    method: 'PUT',
    body: { coordinador_nombre: name },
  });
}
