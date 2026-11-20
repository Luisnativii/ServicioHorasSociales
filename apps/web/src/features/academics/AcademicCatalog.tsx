import { useEffect, useRef } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useSearchParams } from 'react-router-dom';
import { ApiError, NetworkError } from '../../lib/api';
import { getCatalogEntry, listCatalog, type CatalogKind } from './api';

const path = '/catalogos/academico';
const pageSizes = [10, 25, 50];
type Changes = Record<string, string | null>;

function catalogError(error: Error) {
  if (error instanceof NetworkError) return 'No se pudo conectar con el servidor.';
  if (error instanceof ApiError && error.status === 404) return 'El registro ya no está disponible.';
  return 'No se pudo consultar el catálogo.';
}

function EntryState({ active }: { active: boolean }) {
  return <span className={active ? 'account-active' : 'account-inactive'}>{active ? 'Activa' : 'Inactiva'}</span>;
}

export default function AcademicCatalog() {
  const [params, setParams] = useSearchParams();
  const kind: CatalogKind = params.get('vista') === 'carreras' ? 'carreras' : 'facultades';
  const selected = params.get('detalle') ?? '';
  const facultyId = kind === 'carreras' ? params.get('facultad_id') ?? '' : '';
  const search = params.get('q') ?? '';
  const includeInactive = params.get('incluir_inactivas') === 'true';
  const requestedPage = Number(params.get('page') ?? '1');
  const page = Number.isSafeInteger(requestedPage) && requestedPage > 0 ? requestedPage : 1;
  const requestedSize = Number(params.get('page_size') ?? '25');
  const size = pageSizes.includes(requestedSize) ? requestedSize : 25;

  const filters = new URLSearchParams({ page: String(page), page_size: String(size) });
  if (search) filters.set('q', search);
  if (includeInactive) filters.set('incluir_inactivas', 'true');
  if (facultyId) filters.set('facultad_id', facultyId);

  const listing = useQuery({
    queryKey: ['academic-catalog', 'list', kind, filters.toString()],
    queryFn: ({ signal }) => listCatalog(kind, filters, signal),
  });
  const detail = useQuery({
    queryKey: ['academic-catalog', 'detail', kind, selected],
    queryFn: ({ signal }) => getCatalogEntry(kind, selected, signal),
    enabled: Boolean(selected),
  });
  const relatedFacultyId = detail.data && 'facultad_id' in detail.data ? detail.data.facultad_id : '';
  const lookupId = relatedFacultyId || facultyId;
  const faculty = useQuery({
    queryKey: ['academic-catalog', 'detail', 'facultades', lookupId],
    queryFn: ({ signal }) => getCatalogEntry('facultades', lookupId, signal),
    enabled: Boolean(lookupId),
  });
  const heading = useRef<HTMLHeadingElement>(null);
  useEffect(() => {
    if (selected) heading.current?.focus();
  }, [kind, selected]);

  function nextParams(changes: Changes) {
    const next = new URLSearchParams(params);
    for (const [key, value] of Object.entries(changes)) {
      if (value) next.set(key, value);
      else next.delete(key);
    }
    return next;
  }

  function updateFilters(changes: Changes) {
    setParams(nextParams({ page: null, detalle: null, ...changes }));
  }

  function detailLink(id: string) {
    return { pathname: path, search: nextParams({ detalle: id }).toString() };
  }

  function careersLink(id: string, active: boolean) {
    const next = new URLSearchParams({ vista: 'carreras', facultad_id: id });
    if (includeInactive || !active) next.set('incluir_inactivas', 'true');
    return { pathname: path, search: next.toString() };
  }

  const total = listing.data?.pagination.total;
  const pages = total === undefined ? 1 : Math.max(1, Math.ceil(total / size));

  return <section className="accounts-page">
    <h1>Catálogo académico</h1>
    <p className="shell-lead">Consulta las facultades y carreras que mantiene el equipo del core.</p>

    <nav className="academic-views" aria-label="Vista del catálogo">
      <Link to={path} aria-current={kind === 'facultades' ? 'page' : undefined}>Facultades</Link>
      <Link to={`${path}?vista=carreras`} aria-current={kind === 'carreras' ? 'page' : undefined}>Carreras</Link>
    </nav>

    {selected && <section className="academic-detail" aria-labelledby="academic-detail-title">
      <div className="academic-detail-heading">
        <h2 ref={heading} tabIndex={-1} id="academic-detail-title">{detail.data?.nombre ?? 'Detalle del catálogo'}</h2>
        <Link to={{ pathname: path, search: nextParams({ detalle: null }).toString() }}>Cerrar detalle</Link>
      </div>
      {detail.isPending && <p role="status">Cargando detalle…</p>}
      {detail.isError && <div role="alert">
        <p>{catalogError(detail.error)}</p>
        <button type="button" className="shell-secondary" onClick={() => void detail.refetch()}>Reintentar detalle</button>
      </div>}
      {detail.data && <>
        <p>Estado <EntryState active={detail.data.activo} /></p>
        {relatedFacultyId && 'facultad_activa' in detail.data ? <p>
          Facultad {faculty.data?.nombre ?? (faculty.isError ? 'no disponible' : 'cargando…')} <EntryState active={detail.data.facultad_activa} />
        </p> : <Link to={careersLink(detail.data.id, detail.data.activo)}>Ver carreras de esta facultad</Link>}
      </>}
    </section>}

    <form className="accounts-toolbar academic-toolbar" onSubmit={event => {
      event.preventDefault();
      const value = new FormData(event.currentTarget).get('q');
      if (typeof value === 'string') updateFilters({ q: value.trim() });
    }}>
      <label>Nombre de {kind === 'facultades' ? 'la facultad' : 'la carrera'}
        <input key={`${kind}:${search}`} name="q" type="search" defaultValue={search} />
      </label>
      <button type="submit" className="shell-secondary">Buscar</button>
      {search && <button type="button" className="shell-secondary" onClick={() => updateFilters({ q: null })}>Limpiar búsqueda</button>}
      <label className="academic-checkbox">
        <input type="checkbox" checked={includeInactive} onChange={event => updateFilters({
          incluir_inactivas: event.target.checked ? 'true' : null,
        })} /> Incluir inactivas
      </label>
      <label>Resultados por página
        <select value={size} onChange={event => updateFilters({ page_size: event.target.value })}>
          {pageSizes.map(value => <option key={value}>{value}</option>)}
        </select>
      </label>
    </form>

    {facultyId && <div className="academic-faculty-filter">
      <p>Filtrando por facultad{facultyId === lookupId && faculty.data ? ` · ${faculty.data.nombre}` : ''}</p>
      <button type="button" className="shell-secondary" onClick={() => updateFilters({ facultad_id: null })}>Quitar filtro de facultad</button>
    </div>}

    {listing.isPending ? <p role="status" className="accounts-feedback">Cargando {kind}…</p>
      : listing.isError ? <div role="alert" className="accounts-feedback">
        <p>{catalogError(listing.error)}</p>
        <button type="button" className="shell-secondary" onClick={() => void listing.refetch()}>Reintentar listado</button>
      </div>
      : listing.data.entries.length === 0 ? <div className="shell-empty">
        <h2>No hay {kind} en esta página</h2>
        {page > 1 && <button type="button" className="shell-secondary" onClick={() => updateFilters({})}>Primera página</button>}
      </div>
      : <div className="accounts-table-wrap" aria-busy={listing.isFetching}>
        <table className="accounts-table academic-table">
          <caption className="sr-only">{kind === 'facultades' ? 'Facultades' : 'Carreras'} del catálogo académico</caption>
          <thead><tr>
            <th scope="col">Nombre</th><th scope="col">Estado</th>
            {kind === 'facultades' && <th scope="col">Carreras</th>}
          </tr></thead>
          <tbody>{listing.data.entries.map(entry => <tr key={entry.id}>
            <th scope="row"><Link to={detailLink(entry.id)}>{entry.nombre}</Link></th>
            <td><EntryState active={entry.activo} />
              {'facultad_activa' in entry && !entry.facultad_activa && <small>Facultad inactiva</small>}
            </td>
            {kind === 'facultades' && <td><Link to={careersLink(entry.id, entry.activo)} aria-label={`Ver carreras de ${entry.nombre}`}>Ver carreras</Link></td>}
          </tr>)}</tbody>
        </table>
      </div>}

    {total !== undefined && !listing.isError && <div className="accounts-pagination">
      <p role="status">{total} {total === 1 ? 'resultado' : 'resultados'} · Página {page} de {pages}</p>
      <div>
        <button type="button" className="shell-secondary" disabled={page <= 1 || listing.isFetching}
          onClick={() => updateFilters({ page: String(page - 1) })}>Anterior</button>
        <button type="button" className="shell-secondary" disabled={page >= pages || listing.isFetching}
          onClick={() => updateFilters({ page: String(page + 1) })}>Siguiente</button>
      </div>
    </div>}
  </section>;
}
