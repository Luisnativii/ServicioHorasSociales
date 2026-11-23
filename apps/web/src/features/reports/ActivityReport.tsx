import { useState, type FormEvent } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import { useSearchParams } from 'react-router-dom';
import { api, ApiError } from '../../lib/api';
import { listCatalog } from '../academics/api';
import { listClassification } from '../classification/api';
import TerritorySelector, { type TerritorySelection } from '../territory/TerritorySelector';

type Row = {
  carnet: string; estudiante: string; proyecto: string; fecha_horas: string; horas: number;
  naturaleza: string; facultad: string | null; carrera: string | null; tipo: string | null;
  area: string | null; area_css: string | null; distrito: string | null; estado: string;
};
type Report = {
  resumen: {
    estudiantes: number; participaciones: number; proyectos: number;
    internas: number; externas: number; cumplimientos_conocidos: number;
  };
  movimientos: Row[];
  pagina: { total: number };
};
type Option = { id: string; nombre: string };
const size = 25;
const states = ['BORRADOR', 'ACTIVO', 'PAUSADO', 'FINALIZADO', 'CANCELADO', 'ABANDONADO'];
const filterNames = ['desde', 'hasta', 'estado', 'departamento_id', 'municipio_id', 'distrito_id', 'carrera_id', 'tipo_id', 'area_id'];

async function options(kind: 'carreras' | 'tipos-proyecto' | 'areas', signal: AbortSignal) {
  const found: Option[] = [];
  for (let page = 1; ; page++) {
    const params = new URLSearchParams({ page: String(page), page_size: '100' });
    if (kind === 'carreras') {
      params.set('incluir_inactivas', 'true');
      const result = await listCatalog(kind, params, signal);
      found.push(...result.entries);
      if (found.length >= result.pagination.total || !result.entries.length) return found;
    } else {
      params.set('incluir_inactivos', 'true');
      const result = await listClassification(kind, params, signal);
      found.push(...result.opciones);
      if (found.length >= result.pagina.total || !result.opciones.length) return found;
    }
  }
}

function FilterSelect({ field, label, kind, selected, multiple = false }: {
  field: string; label: string; kind: 'carreras' | 'tipos-proyecto' | 'areas'; selected: string[]; multiple?: boolean;
}) {
  const query = useQuery({ queryKey: ['report-options', kind], queryFn: ({ signal }) => options(kind, signal) });
  return <label>{label}
    <select name={field} multiple={multiple} defaultValue={multiple ? selected : selected[0] ?? ''}>
      {!multiple && <option value="">Todas las opciones</option>}
      {selected.filter(id => !query.data?.some(item => item.id === id)).map(id =>
        <option key={id} value={id}>Selección guardada</option>)}
      {query.data?.map(item => <option key={item.id} value={item.id}>{item.nombre}</option>)}
    </select>
    {query.isError && <span role="alert">No se pudo consultar el catálogo.</span>}
  </label>;
}

function Filters({ current, apply, clear }: {
  current: URLSearchParams; apply: (filters: URLSearchParams) => void; clear: () => void;
}) {
  const [territory, setTerritory] = useState<TerritorySelection>({
    departmentId: current.get('departamento_id') ?? '',
    municipalityId: current.get('municipio_id') ?? '',
    districtId: current.get('distrito_id') ?? '',
  });
  const [error, setError] = useState('');
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const next = new URLSearchParams();
    for (const [key, value] of new FormData(event.currentTarget)) if (value) next.append(key, String(value));
    if (territory.departmentId) next.set('departamento_id', territory.departmentId);
    if (territory.municipalityId) next.set('municipio_id', territory.municipalityId);
    if (territory.districtId) next.set('distrito_id', territory.districtId);
    if (next.get('desde') && next.get('hasta') && next.get('desde')! > next.get('hasta')!) {
      setError('La fecha inicial debe ser anterior o igual a la final.');
      return;
    }
    setError('');
    apply(next);
  }
  return <form className="institution-panel institution-form" onSubmit={submit}>
    <label>Actividad desde <input type="date" name="desde" defaultValue={current.get('desde') ?? ''} /></label>
    <label>Actividad hasta <input type="date" name="hasta" defaultValue={current.get('hasta') ?? ''} /></label>
    <label>Estado vigente del proyecto
      <select name="estado" defaultValue={current.get('estado') ?? ''}>
        <option value="">Todos</option>{states.map(state => <option key={state}>{state}</option>)}
      </select>
    </label>
    <fieldset><legend>Geografía vigente del proyecto</legend>
      <TerritorySelector value={territory} onChange={setTerritory} />
    </fieldset>
    <FilterSelect field="carrera_id" label="Carrera histórica" kind="carreras" selected={current.getAll('carrera_id')} />
    <FilterSelect field="tipo_id" label="Tipo histórico" kind="tipos-proyecto" selected={current.getAll('tipo_id')} />
    <FilterSelect field="area_id" label="Áreas históricas" kind="areas" selected={current.getAll('area_id')} multiple />
    {error && <p role="alert">{error}</p>}
    <div className="institution-actions">
      <button type="submit" className="access-primary">Aplicar filtros</button>
      <button type="button" className="shell-secondary" onClick={clear}>Limpiar filtros</button>
    </div>
  </form>;
}

function errorMessage(error: Error) {
  if (error instanceof ApiError && error.status === 413) return 'La descarga supera los 5,000 movimientos. Reduce el intervalo o aplica más filtros.';
  if (error instanceof ApiError && error.status === 400) return 'Revisa los filtros y el intervalo de actividad.';
  return 'No se pudo completar la consulta. Intenta de nuevo.';
}

export default function ActivityReport() {
  const [params, setParams] = useSearchParams();
  const [clearVersion, setClearVersion] = useState(0);
  const filters = new URLSearchParams();
  for (const key of filterNames) for (const value of params.getAll(key)) if (value) filters.append(key, value);
  const filterString = filters.toString();
  const raw = Number(params.get('page') ?? 1);
  const page = Number.isSafeInteger(raw) && raw > 0 ? raw : 1;
  const result = useQuery({
    queryKey: ['activity-report', filterString, page],
    queryFn: ({ signal }) => {
      const url = new URLSearchParams(filterString);
      url.set('page', String(page));
      url.set('page_size', String(size));
      return api.request<Report>(`/api/v1/reportes/actividad?${url}`, { signal });
    },
  });
  const download = useMutation({
    mutationFn: (applied: string) => api.request<Blob>(`/api/v1/reportes/actividad.xlsx?${applied}`, { responseType: 'blob' }),
    onSuccess: blob => {
      const href = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = href;
      link.download = 'actividad-horas-sociales.xlsx';
      link.click();
      setTimeout(() => URL.revokeObjectURL(href), 1000);
    },
  });
  const data = result.data;
  const pages = Math.max(1, Math.ceil((data?.pagina.total ?? 0) / size));
  function changePage(next: number) {
    const query = new URLSearchParams(filterString);
    if (next > 1) query.set('page', String(next));
    setParams(query);
  }

  return <section className="accounts-page">
    <h1>Actividad y reportes</h1>
    <Filters key={`${filterString}:${clearVersion}`} current={filters} apply={next => {
      download.reset(); setParams(next);
    }} clear={() => {
      download.reset(); setClearVersion(version => version + 1); setParams(new URLSearchParams());
    }} />
    {result.isPending && <p role="status">Cargando actividad…</p>}
    {result.isError && <div role="alert">{errorMessage(result.error)}{' '}
      <button type="button" className="shell-secondary" onClick={() => void result.refetch()}>Reintentar</button></div>}
    {data && !result.isError && <>
      <dl className="institution-facts">
        <div><dt>Estudiantes únicos</dt><dd>{data.resumen.estudiantes}</dd></div>
        <div><dt>Participaciones</dt><dd>{data.resumen.participaciones}</dd></div>
        <div><dt>Proyectos</dt><dd>{data.resumen.proyectos}</dd></div>
        <div><dt>Horas internas</dt><dd>{data.resumen.internas}</dd></div>
        <div><dt>Horas externas</dt><dd>{data.resumen.externas}</dd></div>
        <div><dt>Cumplimientos con fecha conocida</dt><dd>{data.resumen.cumplimientos_conocidos}</dd></div>
      </dl>
      <p>El cumplimiento se cuenta cuando el libro completo del estudiante alcanza por primera vez 300 horas de cada naturaleza.</p>
      <button type="button" className="shell-secondary" disabled={download.isPending}
        onClick={() => download.mutate(filterString)}>{download.isPending ? 'Preparando Excel…' : 'Descargar Excel'}</button>
      {download.error && <p role="alert">{errorMessage(download.error)}</p>}
      {data.movimientos.length ? <div className="accounts-table-wrap"><table className="accounts-table">
        <thead><tr><th scope="col">Estudiante</th><th scope="col">Proyecto</th><th scope="col">Actividad</th><th scope="col">Horas</th>
          <th scope="col">Academia histórica</th><th scope="col">Clasificación histórica</th><th scope="col">Referencias vigentes</th>
        </tr></thead>
        <tbody>{data.movimientos.map((row, idx) => <tr key={idx}>
          <th scope="row">{row.estudiante}<small>{row.carnet}</small></th><td>{row.proyecto}</td><td>{row.fecha_horas}</td>
          <td>{row.horas} · {row.naturaleza}</td>
          <td>{row.carrera ?? 'Sin dato histórico'}{row.facultad && <small>{row.facultad}</small>}</td>
          <td>{row.tipo ?? 'Sin dato histórico'} · {row.area ?? 'Sin dato histórico'}{row.area_css && <small>Área CSS original: {row.area_css}</small>}</td>
          <td>{row.distrito ?? 'Sin distrito'} · {row.estado}</td>
        </tr>)}</tbody>
      </table></div> : <p>No hay movimientos en esta página.</p>}
      <nav className="accounts-pagination" aria-label="Páginas del informe"><p>{data.pagina.total} movimientos · Página {page} de {pages}</p>
        <div><button type="button" className="shell-secondary" disabled={page <= 1} onClick={() => changePage(Math.min(page - 1, pages))}>Anterior</button>
        <button type="button" className="shell-secondary" disabled={page >= pages} onClick={() => changePage(page + 1)}>Siguiente</button></div>
      </nav>
    </>}
  </section>;
}
