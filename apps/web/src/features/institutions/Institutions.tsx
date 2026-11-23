import { useEffect, useRef } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link, useSearchParams } from 'react-router-dom';
import { deactivateInstitution, getInstitution, institutionKey, listInstitutions, type Institution } from './api';
import InstitutionForm, { institutionError } from './InstitutionForm';

const pageSize = 25;
const maxPage = Math.floor(0x7fffffff / pageSize) + 1;

function QueryFailure({ error, retry }: { error: Error; retry: () => void }) {
  return <div className="accounts-feedback">
    <p role="alert">{institutionError(error)}</p>
    <button type="button" className="shell-secondary" onClick={retry}>Volver a consultar</button>
  </div>;
}

function InstitutionDetail({ institution, edit }: { institution: Institution; edit: () => void }) {
  const dialog = useRef<HTMLDialogElement>(null);
  const client = useQueryClient();
  const deactivate = useMutation({
    mutationFn: () => deactivateInstitution(institution.id),
    onSuccess: result => {
      dialog.current?.close();
      void client.invalidateQueries({ queryKey: institutionKey });
      client.setQueryData([...institutionKey, 'detail', result.id], result);
    },
  });

  return <section className="institution-panel" aria-labelledby="institution-detail-title">
    <header className="institution-heading">
      <h2 id="institution-detail-title" tabIndex={-1}>{institution.nombre}</h2>
      <span role="status">{institution.activo ? 'Activa' : 'Inactiva'}</span>
    </header>
    <dl className="institution-facts">
      <div><dt>Código CSS</dt><dd>{institution.codigo_css ?? 'Sin código'}</dd></div>
      <div><dt>Distrito de la sede</dt><dd>{institution.sede?.nombre ?? 'Sin identificar'}</dd></div>
      {institution.ubicacion_heredada && <div><dt>Ubicación histórica</dt><dd>{institution.ubicacion_heredada}</dd></div>}
    </dl>
    <h3>Contacto principal</h3>
    {institution.contacto ? <dl className="institution-facts">
      <div><dt>Nombre</dt><dd>{institution.contacto.nombre ?? 'Sin registrar'}</dd></div>
      <div><dt>Correo</dt><dd>{institution.contacto.correo ?? 'Sin registrar'}</dd></div>
      <div><dt>Teléfono</dt><dd>{institution.contacto.telefono ?? 'Sin registrar'}</dd></div>
      <div><dt>Extensión</dt><dd>{institution.contacto.extension ?? 'Sin registrar'}</dd></div>
    </dl> : <p>Sin contacto registrado.</p>}
    <h3>Proyectos asociados</h3>
    <p>La consulta de proyectos asociados aún no está disponible.</p>
    <div className="institution-actions">
      <button type="button" className="shell-secondary" onClick={edit}>Editar institución</button>
      {institution.activo && <button type="button" className="shell-secondary" onClick={() => {
        deactivate.reset();
        dialog.current?.showModal();
      }}>Desactivar institución</button>}
    </div>
    <dialog ref={dialog} className="shell-confirm" aria-labelledby="institution-confirm-title"
      onCancel={event => { if (deactivate.isPending) event.preventDefault(); }}
      onClose={() => document.getElementById('institution-detail-title')?.focus()}>
      <h2 id="institution-confirm-title">¿Desactivar esta institución?</h2>
      <p>{institution.nombre} seguirá disponible en el listado de inactivas con sus datos históricos.</p>
      {deactivate.error && <p className="access-error" role="alert">{institutionError(deactivate.error)}</p>}
      <div className="shell-dialog-actions">
        <button type="button" className="shell-secondary" autoFocus disabled={deactivate.isPending}
          onClick={() => dialog.current?.close()}>Cancelar</button>
        <button type="button" className="access-primary" disabled={deactivate.isPending}
          onClick={() => deactivate.mutate()}>{deactivate.isPending ? 'Desactivando…' : 'Desactivar'}</button>
      </div>
    </dialog>
  </section>;
}

export default function Institutions() {
  const [params, setParams] = useSearchParams();
  const id = params.get('institucion') ?? '';
  const view = params.get('vista');
  const creating = view === 'alta';
  const editing = id !== '' && view === 'editar';
  const showList = !creating && !id;
  const pageValue = Number(params.get('page') ?? 1);
  const page = Number.isSafeInteger(pageValue) && pageValue > 0 && pageValue <= maxPage ? pageValue : 1;
  const search = [...(params.get('q') ?? '')].slice(0, 200).join('');
  const includeInactive = params.get('incluir_inactivas') === 'true';

  function changeParams(changes: Record<string, string | null>) {
    const next = new URLSearchParams(params);
    for (const [key, value] of Object.entries(changes)) {
      if (value) next.set(key, value);
      else next.delete(key);
    }
    return next;
  }

  const filters = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
  if (search) filters.set('q', search);
  if (includeInactive) filters.set('incluir_inactivas', 'true');

  const listing = useQuery({
    queryKey: [...institutionKey, 'list', filters.toString()],
    queryFn: ({ signal }) => listInstitutions(filters, signal),
    enabled: showList,
  });
  const detail = useQuery({
    queryKey: [...institutionKey, 'detail', id],
    queryFn: ({ signal }) => getInstitution(id, signal),
    enabled: Boolean(id) && !creating,
  });
  const pages = Math.max(1, Math.ceil((listing.data?.pagina.total ?? 0) / pageSize));
  const heading = useRef<HTMLHeadingElement>(null);

  useEffect(() => {
    const title = creating || editing ? 'institution-form-title' : 'institution-detail-title';
    (document.getElementById(title) ?? heading.current)?.focus();
  }, [id, creating, editing, Boolean(detail.data)]);

  function showDetail(result: Institution) {
    setParams(changeParams({ institucion: result.id, vista: null }), { replace: true });
  }

  return <section className="accounts-page">
    <h1 ref={heading} tabIndex={-1}>Instituciones</h1>
    {!showList && <Link className="institution-back shell-text-link"
      to={{ search: changeParams({ institucion: null, vista: null }).toString() }}>Volver al listado</Link>}
    {creating && <InstitutionForm key="new" onSaved={showDetail}
      onCancel={() => setParams(changeParams({ institucion: null, vista: null }), { replace: true })} />}
    {!creating && id && <>
      {detail.isPending && <p className="accounts-feedback" role="status">Cargando institución…</p>}
      {detail.isError && <QueryFailure error={detail.error} retry={() => void detail.refetch()} />}
      {detail.data && (editing
        ? <InstitutionForm key={id} institution={detail.data} onSaved={showDetail}
          onCancel={() => setParams(changeParams({ vista: null }), { replace: true })} />
        : <InstitutionDetail key={id} institution={detail.data}
          edit={() => setParams(changeParams({ vista: 'editar' }))} />)}
    </>}
    {showList && <>
      <form className="accounts-toolbar academic-toolbar" onSubmit={event => {
        event.preventDefault();
        const query = String(new FormData(event.currentTarget).get('q') ?? '').trim();
        setParams(changeParams({ q: query, page: null }));
      }}>
        <label>Buscar por nombre o código CSS
          <input type="search" name="q" key={search} defaultValue={search} maxLength={200} />
        </label>
        <button className="shell-secondary" type="submit">Buscar</button>
        {search && <button className="shell-secondary" type="button"
          onClick={() => setParams(changeParams({ q: null, page: null }))}>Limpiar búsqueda</button>}
        <label className="academic-checkbox">
          <input type="checkbox" checked={includeInactive} onChange={event =>
            setParams(changeParams({ incluir_inactivas: event.target.checked ? 'true' : null, page: null }))} />
          Incluir inactivas
        </label>
        <button className="shell-secondary" type="button"
          onClick={() => setParams(changeParams({ vista: 'alta' }))}>Nueva institución</button>
      </form>
      {listing.isPending && <p className="accounts-feedback" role="status">Cargando instituciones…</p>}
      {listing.isError && <QueryFailure error={listing.error} retry={() => void listing.refetch()} />}
      {listing.data && <>
        {listing.data.instituciones.length === 0 ? <p className="accounts-feedback" role="status">No hay instituciones en esta página.</p>
          : <div className="accounts-table-wrap">
            <table className="accounts-table academic-table" aria-label="Instituciones">
              <thead><tr><th scope="col">Nombre</th><th scope="col">Sede</th><th scope="col">Estado</th></tr></thead>
              <tbody>{listing.data.instituciones.map(item => <tr key={item.id}>
                <th scope="row">
                  <Link to={{ search: changeParams({ institucion: item.id }).toString() }}>{item.nombre}</Link>
                  {item.codigo_css && <small>{item.codigo_css}</small>}
                </th>
                <td>{item.sede?.nombre ?? 'Sin identificar'}</td>
                <td className={item.activo ? 'account-active' : 'account-inactive'}>{item.activo ? 'Activa' : 'Inactiva'}</td>
              </tr>)}</tbody>
            </table>
          </div>}
        <nav className="accounts-pagination" aria-label="Páginas de instituciones">
          <p aria-live="polite">{listing.data.pagina.total} resultados · Página {page} de {pages}</p>
          <div>
            <button className="shell-secondary" type="button" disabled={page <= 1}
              onClick={() => setParams(changeParams({ page: String(Math.min(page - 1, pages)) }))}>Anterior</button>
            <button className="shell-secondary" type="button" disabled={page >= pages}
              onClick={() => setParams(changeParams({ page: String(page + 1) }))}>Siguiente</button>
          </div>
        </nav>
      </>}
    </>}
  </section>;
}
