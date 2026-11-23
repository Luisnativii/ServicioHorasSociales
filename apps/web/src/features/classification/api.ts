import { api } from '../../lib/api';

export type ClassificationKind = 'tipos-proyecto' | 'areas';
export type ClassificationOption = { id: string; nombre: string; activo: boolean };
type ClassificationPage = {
  opciones: ClassificationOption[];
  pagina: { page: number; page_size: number; total: number };
};

export type ClassificationChange =
  | { action: 'create'; name: string }
  | { action: 'edit'; id: string; name: string }
  | { action: 'deactivate'; id: string };

const base = '/api/v1/catalogos';

export function listClassification(kind: ClassificationKind, params: URLSearchParams, signal: AbortSignal) {
  return api.request<ClassificationPage>(`${base}/${kind}?${params}`, { signal });
}

export function changeClassification(kind: ClassificationKind, change: ClassificationChange) {
  const collection = `${base}/${kind}`;
  if (change.action === 'create') {
    return api.request<ClassificationOption>(collection, { method: 'POST', body: { nombre: change.name } });
  }
  const item = `${collection}/${encodeURIComponent(change.id)}`;
  if (change.action === 'edit') {
    return api.request<ClassificationOption>(item, { method: 'PATCH', body: { nombre: change.name } });
  }
  return api.request<ClassificationOption>(`${item}/desactivar`, { method: 'POST' });
}
