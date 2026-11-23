import { useState, type FormEvent } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link, useSearchParams } from 'react-router-dom';
import { api, ApiError } from '../../lib/api';
import { listClassification, type ClassificationKind, type ClassificationOption } from '../classification/api';
import { listInstitutions, type Institution } from '../institutions/api';
import TerritorySelector from '../territory/TerritorySelector';

type Contact = { nombre: string; correo: string | null; telefono: string | null; extension: string | null };
type Project = {
  id: string; titulo: string; descripcion: string; estado: string; naturaleza: string; cupo: number | null;
  fecha_inicio: string | null; fecha_cierre: string | null; correlativo_css: string | null;
  institucion_id: string; institucion: string; sede: string | null; distrito_id: string | null;
  distrito: string | null; ubicacion_heredada: string | null; area_css: string | null;
  tipo_id: string | null; tipo: string | null; area_id: string | null; area: string | null;
  modalidad_id: string | null; modalidad: string | null; responsable: Contact | null;
};
type Page = { proyectos: Project[]; pagina: { total: number } };
const base = '/api/v1/proyectos';
const size = 25;
const emptyTerritory = { departmentId: '', municipalityId: '', districtId: '' };
const catalogFields = [
  { id: 'tipo_id', name: 'tipo', label: 'Tipo de proyecto', kind: 'tipos-proyecto' },
  { id: 'area_id', name: 'area', label: 'Área', kind: 'areas' },
  { id: 'modalidad_id', name: 'modalidad', label: 'Modalidad', kind: 'modalidades' },
] as const;

async function options(kind: ClassificationKind, signal: AbortSignal, includeInactive = false) {
  const values: ClassificationOption[] = [];
  for (let page = 1; ; page++) {
    const result = await listClassification(kind, new URLSearchParams({ page: String(page), page_size: '100', incluir_inactivos: String(includeInactive) }), signal);
    values.push(...result.opciones);
    if (values.length >= result.pagina.total || !result.opciones.length) return values;
  }
}

async function institutions(signal: AbortSignal) {
  const values: Institution[] = [];
  for (let page = 1; ; page++) {
    const result = await listInstitutions(new URLSearchParams({ page: String(page), page_size: '100', incluir_inactivas: 'true' }), signal);
    values.push(...result.instituciones);
    if (values.length >= result.pagina.total || !result.instituciones.length) return values;
  }
}

function message(error: Error) {
  if (error instanceof ApiError) {
    if (error.status === 404) return 'El proyecto no está disponible.';
    if (error.status === 409) return 'Ya existe un proyecto con ese título en la institución.';
    if (error.field === 'escritura_core') return 'La edición de proyectos todavía no está habilitada en este entorno.';
    if (error.status === 400) return `Revisa el campo ${error.field?.replaceAll('_', ' ') ?? 'indicado'}.`;
  }
  return 'No se pudo completar la operación. Intenta de nuevo.';
}

function CatalogSelect({ item, project }: { item: typeof catalogFields[number]; project?: Project }) {
  const query = useQuery({ queryKey: ['project-options', item.kind], queryFn: ({ signal }) => options(item.kind, signal) });
  const current = project?.[item.id] ?? '';
  return <label>{item.label}
    <select name={item.id} defaultValue={current} required>
      <option value="">Seleccionar</option>
      {current && !query.data?.some(option => option.id === current) && <option value={current}>{project?.[item.name]}</option>}
      {query.data?.map(option => <option key={option.id} value={option.id}>{option.nombre}</option>)}
    </select>
    {query.isError && <span role="alert">No se pudo consultar el catálogo.</span>}
    {query.data?.length === 0 && !current && <Link to={`/catalogos/clasificacion?vista=${item.kind}`}>Agregar una opción al catálogo</Link>}
  </label>;
}

function ProjectForm({ project, saved, cancel }: { project?: Project; saved: (p: Project) => void; cancel: () => void }) {
  const [selected, select] = useState(project?.institucion_id ?? '');
  const [district, setDistrict] = useState(emptyTerritory);
  const [changeDistrict, setChangeDistrict] = useState(!project?.distrito_id);
  const [error, setError] = useState('');
  const orgs = useQuery({ queryKey: ['project-institutions'], queryFn: ({ signal }) => institutions(signal) });
  const org = orgs.data?.find(item => item.id === selected);
  const save = useMutation({
    mutationFn: (body: unknown) => api.request<Project>(project ? `${base}/${project.id}` : base, {
      method: project ? 'PUT' : 'POST', body,
    }), onSuccess: saved,
  });
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (save.isPending) return;
    const data = new FormData(event.currentTarget);
    const value = (name: string) => String(data.get(name) ?? '').trim();
    const distrito_id = changeDistrict ? district.districtId : project?.distrito_id;
    if (!distrito_id) { setError('Selecciona el distrito de ejecución.'); return; }
    setError('');
    save.mutate({
      titulo: value('titulo'), descripcion: value('descripcion'), institucion_id: selected, distrito_id,
      naturaleza: value('naturaleza'), cupo: value('cupo') ? Number(value('cupo')) : null,
      fecha_inicio: value('fecha_inicio') || null, tipo_id: value('tipo_id'), area_id: value('area_id'),
      modalidad_id: value('modalidad_id'), responsable: {
        nombre: value('nombre'), correo: value('correo') || null, telefono: value('telefono') || null, extension: value('extension') || null,
      },
    });
  }
  return <section className="institution-panel"><h2>{project ? 'Editar proyecto' : 'Nuevo proyecto'}</h2>
    <form className="institution-form" onSubmit={submit}>
      <label>Título <input name="titulo" defaultValue={project?.titulo} maxLength={300} required /></label>
      <label>Descripción <textarea name="descripcion" defaultValue={project?.descripcion} maxLength={10000} /></label>
      {project ? <p>Institución: {project.institucion}</p> : <label>Institución
        <select required value={selected} onChange={event => select(event.target.value)}>
          <option value="">Seleccionar</option>
          {orgs.data?.filter(item => item.activo).map(item => <option key={item.id} value={item.id}>{item.nombre}</option>)}
        </select>
      </label>}
      {orgs.isError && !project && <p role="alert">No se pudieron cargar las instituciones.</p>}
      <p>Sede institucional: {org?.sede?.nombre ?? project?.sede ?? 'Sin identificar'}</p>
      <fieldset><legend>Lugar de ejecución</legend>
        {project?.distrito_id && <label><input type="checkbox" checked={changeDistrict} onChange={event => setChangeDistrict(event.target.checked)} /> Cambiar distrito guardado ({project.distrito})</label>}
        {changeDistrict && <TerritorySelector value={district} onChange={setDistrict} historicalName={project?.ubicacion_heredada} />}
      </fieldset>
      <label>Naturaleza <select name="naturaleza" defaultValue={project?.naturaleza ?? 'EXTERNA'}><option>EXTERNA</option><option>INTERNA</option></select></label>
      <label>Fecha de inicio <input name="fecha_inicio" type="date" defaultValue={project?.fecha_inicio ?? ''} /></label>
      <label>Cupo solicitado <input name="cupo" type="number" step={1} min={1} max={2147483647} defaultValue={project?.cupo ?? ''} /></label>
      {catalogFields.map(item => <CatalogSelect key={item.id} item={item} project={project} />)}
      <fieldset><legend>Responsable del proyecto</legend>
        <label>Nombre <input name="nombre" maxLength={200} defaultValue={project?.responsable?.nombre} required /></label>
        <label>Correo <input name="correo" type="email" defaultValue={project?.responsable?.correo ?? ''} /></label>
        <label>Teléfono <input name="telefono" type="tel" maxLength={20} defaultValue={project?.responsable?.telefono ?? ''} /></label>
        <label>Extensión <input name="extension" maxLength={200} defaultValue={project?.responsable?.extension ?? ''} /></label>
      </fieldset>
      {(error || save.error) && <p role="alert" className="access-error">{error || (save.error && message(save.error))}</p>}
      <div className="institution-actions"><button type="button" className="shell-secondary" disabled={save.isPending} onClick={cancel}>Cancelar</button>
      <button type="submit" className="access-primary" disabled={save.isPending}>{save.isPending ? 'Guardando…' : 'Guardar'}</button></div>
    </form>
  </section>;
}

function Detail({ project, edit }: { project: Project; edit: () => void }) {
  const facts = [
    ['Institución', project.institucion], ['Sede institucional', project.sede], ['Distrito de ejecución', project.distrito],
    ['Ubicación histórica', project.ubicacion_heredada], ['Correlativo CSS', project.correlativo_css], ['Área histórica CSS', project.area_css],
    ['Tipo', project.tipo], ['Área', project.area], ['Modalidad', project.modalidad],
    ['Naturaleza', project.naturaleza], ['Estado', project.estado], ['Cupo solicitado', project.cupo],
    ['Inicio', project.fecha_inicio], ['Cierre', project.fecha_cierre], ['Responsable', project.responsable?.nombre],
    ['Correo', project.responsable?.correo], ['Teléfono', project.responsable?.telefono], ['Extensión', project.responsable?.extension],
  ];
  return <section className="institution-panel"><h2>{project.titulo}</h2><p>{project.descripcion}</p>
    <dl className="institution-facts">{facts.map(([label, value]) => <div key={label}><dt>{label}</dt><dd>{value ?? 'Sin registrar'}</dd></div>)}</dl>
    <button type="button" className="shell-secondary" onClick={edit}>Editar proyecto</button>
  </section>;
}

export default function Projects() {
  const [params, setParams] = useSearchParams();
  const client = useQueryClient();
  const id = params.get('proyecto') ?? '';
  const view = params.get('vista');
  const creating = view === 'alta';
  const listing = !creating && !id;
  const rawPage = Number(params.get('page') ?? 1);
  const page = Number.isSafeInteger(rawPage) && rawPage > 0 ? rawPage : 1;
  const filters = new URLSearchParams({ page: String(page), page_size: String(size) });
  for (const field of ['q', 'estado', 'institucion_id', 'area_id']) {
    for (const value of params.getAll(field)) if (value) filters.append(field, value);
  }
  const list = useQuery({ queryKey: ['projects', 'list', filters.toString()], enabled: listing,
    queryFn: ({ signal }) => api.request<Page>(`${base}?${filters}`, { signal }) });
  const detail = useQuery({ queryKey: ['projects', 'detail', id], enabled: !!id && !creating,
    queryFn: ({ signal }) => api.request<Project>(`${base}/${encodeURIComponent(id)}`, { signal }) });
  const areas = useQuery({ queryKey: ['project-options', 'area-filter'], enabled: listing, queryFn: ({ signal }) => options('areas', signal, true) });
  const orgs = useQuery({ queryKey: ['project-institutions'], enabled: listing, queryFn: ({ signal }) => institutions(signal) });
  const pages = Math.max(1, Math.ceil((list.data?.pagina.total ?? 0) / size));
  function next(changes: Record<string, string | string[]>) {
    const updated = new URLSearchParams(params);
    for (const [key, value] of Object.entries(changes)) {
      updated.delete(key);
      for (const item of Array.isArray(value) ? value : [value]) if (item) updated.append(key, item);
    }
    return updated;
  }
  function saved(project: Project) {
    client.setQueryData(['projects', 'detail', project.id], project);
    void client.invalidateQueries({ queryKey: ['projects', 'list'] });
    setParams(next({ proyecto: project.id, vista: '' }), { replace: true });
  }
  return <section className="accounts-page"><h1>Proyectos</h1>
    {!listing && <Link to={{ search: next({ proyecto: '', vista: '' }).toString() }}>Volver al listado</Link>}
    {creating && <ProjectForm key="new" saved={saved} cancel={() => setParams(next({ vista: '' }))} />}
    {!creating && id && <>
      {detail.isPending && <p role="status">Cargando proyecto…</p>}
      {detail.isError ? <p role="alert">{message(detail.error)}</p> : detail.data && (view === 'editar'
        ? <ProjectForm key={id} project={detail.data} saved={saved} cancel={() => setParams(next({ vista: '' }))} />
        : <Detail project={detail.data} edit={() => setParams(next({ vista: 'editar' }))} />)}
    </>}
    {listing && <>
      <form key={filters.toString()} className="accounts-toolbar academic-toolbar" onSubmit={event => {
        event.preventDefault();
        const form = new FormData(event.currentTarget);
        setParams(next({ q: String(form.get('q') ?? '').trim(), estado: String(form.get('estado') ?? ''),
          institucion_id: String(form.get('institucion_id') ?? ''), area_id: form.getAll('area_id').map(String), page: '' }));
      }}>
        <label>Nombre, institución o CSS <input name="q" type="search" maxLength={200} defaultValue={params.get('q') ?? ''} /></label>
        <label>Estado <select name="estado" defaultValue={params.get('estado') ?? ''}><option value="">Todos</option>
          {['BORRADOR', 'ACTIVO', 'PAUSADO', 'FINALIZADO', 'CANCELADO', 'ABANDONADO'].map(state => <option key={state}>{state}</option>)}</select></label>
        <label>Institución <select name="institucion_id" defaultValue={params.get('institucion_id') ?? ''}><option value="">Todas</option>
          {params.get('institucion_id') && !orgs.data?.some(org => org.id === params.get('institucion_id')) && <option value={params.get('institucion_id')!}>Institución seleccionada</option>}
          {orgs.data?.map(org => <option key={org.id} value={org.id}>{org.nombre}</option>)}</select></label>
        <label>Áreas <select multiple name="area_id" defaultValue={params.getAll('area_id')}>
          {params.getAll('area_id').filter(id => !areas.data?.some(area => area.id === id)).map(id => <option key={id} value={id}>Área seleccionada</option>)}
          {areas.data?.map(area => <option key={area.id} value={area.id}>{area.nombre}</option>)}</select></label>
        <button type="submit" className="shell-secondary">Buscar</button>
        <button type="button" className="shell-secondary" onClick={() => setParams(next({ vista: 'alta' }))}>Nuevo proyecto</button>
      </form>
      {list.isPending && <p role="status">Cargando proyectos…</p>}
      {list.isError ? <p role="alert">{message(list.error)}</p> : list.data && <>
        {list.data.proyectos.length ? <div className="accounts-table-wrap"><table className="accounts-table academic-table" aria-label="Proyectos">
          <thead><tr><th scope="col">Proyecto</th><th scope="col">Institución</th><th scope="col">Estado</th></tr></thead>
          <tbody>{list.data.proyectos.map(project => <tr key={project.id}>
            <th scope="row"><Link to={{ search: next({ proyecto: project.id }).toString() }}>{project.titulo}</Link></th>
            <td>{project.institucion}</td><td>{project.estado}</td>
          </tr>)}</tbody>
        </table></div> : <p role="status">No hay proyectos en esta página.</p>}
        <nav className="accounts-pagination" aria-label="Páginas de proyectos"><p>{list.data.pagina.total} resultados · Página {page} de {pages}</p><div>
          <button className="shell-secondary" type="button" disabled={page <= 1} onClick={() => setParams(next({ page: String(Math.min(pages, page - 1)) }))}>Anterior</button>
          <button className="shell-secondary" type="button" disabled={page >= pages} onClick={() => setParams(next({ page: String(page + 1) }))}>Siguiente</button>
        </div></nav>
      </>}
    </>}
  </section>;
}
