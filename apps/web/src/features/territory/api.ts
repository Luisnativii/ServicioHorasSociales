import { api } from '../../lib/api';

export type TerritoryLevel = 'departamentos' | 'municipios' | 'distritos';
export type TerritoryPlace = { id: string; codigo: string; nombre: string; padre_id?: string };

type TerritoryPage = {
  lugares: TerritoryPlace[];
  pagina: { page: number; page_size: number; total: number };
};

export async function listTerritory(level: TerritoryLevel, parentId: string, signal: AbortSignal) {
  const places: TerritoryPlace[] = [];
  const params = new URLSearchParams({ page_size: '100' });
  if (level === 'municipios') params.set('departamento_id', parentId);
  if (level === 'distritos') params.set('municipio_id', parentId);

  for (let page = 1; ; page++) {
    params.set('page', String(page));
    const result = await api.request<TerritoryPage>(`/api/v1/catalogos/${level}?${params}`, { signal });
    places.push(...result.lugares.filter(place => (place.padre_id ?? '') === parentId));
    if (result.lugares.length === 0 || page * result.pagina.page_size >= result.pagina.total) {
      return places;
    }
  }
}
