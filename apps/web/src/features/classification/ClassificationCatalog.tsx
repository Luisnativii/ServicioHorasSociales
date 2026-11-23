import { useEffect, useRef, useState, type FormEvent } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link, useSearchParams } from 'react-router-dom';
import { ApiError, NetworkError } from '../../lib/api';
import { changeClassification, listClassification, type ClassificationChange, type ClassificationKind, type ClassificationOption } from './api';

const path = '/catalogos/clasificacion';
const pageSizes = [10, 25, 50];
type Editor = { action: 'create' } | { action: 'edit' | 'deactivate'; option: ClassificationOption };

function errorMessage(error: Error) {
  if (error instanceof NetworkError) return 'Se perdió la conexión. Actualiza el listado para comprobar si el cambio se guardó.';
  if (error instanceof ApiError && error.status === 409) return 'Ya existe una opción con ese nombre en este catálogo.';
  if (error instanceof ApiError && error.status === 400) return 'Revisa el nombre e inténtalo de nuevo.';
  if (error instanceof ApiError && error.status === 404) return 'La opción ya no está disponible.';
  return 'No se pudo completar el cambio.';
}

export default function ClassificationCatalog() {
  const [params] = useSearchParams();
  const kind = params.get('vista') === 'areas' ? 'areas' : 'tipos-proyecto';
  return <section className="accounts-page">
    <h1>Tipos y áreas de proyecto</h1>
    <nav className="academic-views" aria-label="Catálogo de clasificación">
      <Link to={path} aria-current={kind === 'tipos-proyecto' ? 'page' : undefined}>Tipos de proyecto</Link>
      <Link to={`${path}?vista=areas`} aria-current={kind === 'areas' ? 'page' : undefined}>Áreas</Link>
    </nav>
    <ClassificationList key={kind} kind={kind} />
  </section>;
}

function ClassificationList({ kind }: { kind: ClassificationKind }) {
  const [params, setParams] = useSearchParams();
  const client = useQueryClient();
  const dialog = useRef<HTMLDialogElement>(null);
  const opener = useRef<HTMLButtonElement>(null);
  const createButton = useRef<HTMLButtonElement>(null);
  const submitting = useRef(false);
  const [editor, setEditor] = useState<Editor | null>(null);
  const [name, setName] = useState('');
  const [validation, setValidation] = useState('');
  const [notice, setNotice] = useState('');
  const pageValue = Number(params.get('page') ?? '1');
  const page = Number.isSafeInteger(pageValue) && pageValue > 0 ? pageValue : 1;
  const sizeValue = Number(params.get('page_size') ?? '25');
  const size = pageSizes.includes(sizeValue) ? sizeValue : 25;
  const search = [...(params.get('q') ?? '')].slice(0, 200).join('');
  const includeInactive = params.get('incluir_inactivos') === 'true';
  const filters = new URLSearchParams({ page: String(page), page_size: String(size) });
  if (search) filters.set('q', search);
  if (includeInactive) filters.set('incluir_inactivos', 'true');

  const queryKey = ['classification-catalog', kind];
  const listing = useQuery({
    queryKey: [...queryKey, filters.toString()],
    queryFn: ({ signal }) => listClassification(kind, filters, signal),
  });
  const mutation = useMutation({
    mutationFn: (change: ClassificationChange) => changeClassification(kind, change),
    onSuccess: async (_, change) => {
      setNotice(change.action === 'deactivate' ? 'Opción desactivada.' : 'Nombre guardado.');
      await client.invalidateQueries({ queryKey });
      closeEditor();
    },
    onError: error => {
      if (error instanceof NetworkError || (error instanceof ApiError && error.status >= 500)) {
        void client.invalidateQueries({ queryKey });
      }
    },
    onSettled: () => { submitting.current = false; },
  });

  useEffect(() => {
    if (editor && !dialog.current?.open) dialog.current?.showModal();
  }, [editor]);

  function updateFilters(changes: Record<string, string | null>) {
    setNotice('');
    const next = new URLSearchParams(params);
    for (const [key, value] of Object.entries({ page: null, ...changes })) {
      if (value) next.set(key, value);
      else next.delete(key);
    }
    setParams(next);
  }

  function closeEditor() {
    setEditor(null);
    dialog.current?.close();
  }

  function open(next: Editor, button: HTMLButtonElement) {
    opener.current = button;
    setName(next.action === 'create' ? '' : next.option.nombre);
    setValidation('');
    setNotice('');
    mutation.reset();
    setEditor(next);
  }

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!editor || submitting.current) return;
    const trimmed = name.trim();
    if (editor.action !== 'deactivate' && (!trimmed || [...trimmed].length > 200 || trimmed.includes('\u0000'))) {
      setValidation('Escribe un nombre de entre 1 y 200 caracteres.');
      return;
    }
    setValidation('');
    submitting.current = true;
    if (editor.action === 'create') mutation.mutate({ action: 'create', name: trimmed });
    else if (editor.action === 'edit') mutation.mutate({ action: 'edit', id: editor.option.id, name: trimmed });
    else mutation.mutate({ action: 'deactivate', id: editor.option.id });
  }

  const isArea = kind === 'areas';
  const label = isArea ? 'área' : 'tipo de proyecto';
  const plural = isArea ? 'áreas' : 'tipos de proyecto';
  const total = listing.data?.pagina.total ?? 0;
  const pages = Math.max(1, Math.ceil(total / size));

  return <>
    <form className="accounts-toolbar academic-toolbar" onSubmit={event => {
      event.preventDefault();
      const value = new FormData(event.currentTarget).get('q');
      if (typeof value === 'string') updateFilters({ q: value.trim() });
    }}>
      <label>Buscar por nombre
        <input type="search" name="q" key={search} defaultValue={search} maxLength={200} />
      </label>
      <button type="submit" className="shell-secondary">Buscar</button>
      {search && <button type="button" className="shell-secondary" onClick={() => updateFilters({ q: null })}>Limpiar búsqueda</button>}
      <label className="academic-checkbox">
        <input type="checkbox" checked={includeInactive}
          onChange={event => updateFilters({ incluir_inactivos: event.target.checked ? 'true' : null })} />
        Incluir inactivos
      </label>
      <label>Resultados por página
        <select value={size} onChange={event => updateFilters({ page_size: event.target.value })}>
          {pageSizes.map(value => <option key={value}>{value}</option>)}
        </select>
      </label>
      <button ref={createButton} type="button" className="shell-secondary"
        onClick={event => open({ action: 'create' }, event.currentTarget)}>{isArea ? 'Nueva área' : 'Nuevo tipo'}</button>
    </form>

    {notice && <p role="status" className="accounts-success">{notice}</p>}
    {listing.isPending ? <p role="status" className="accounts-feedback">Cargando {plural}…</p>
      : listing.isError ? <div role="alert" className="accounts-feedback">
        <p>No se pudo consultar el catálogo.</p>
        <button type="button" className="shell-secondary" onClick={() => void listing.refetch()}>Reintentar listado</button>
      </div>
      : listing.data.opciones.length === 0 ? <div className="shell-empty">
        <h2>{search || page > 1 ? 'No hay resultados en esta página' : `No hay ${plural}${includeInactive ? '' : isArea ? ' activas' : ' activos'}`}</h2>
        {page > 1 && <button type="button" className="shell-secondary" onClick={() => updateFilters({})}>Primera página</button>}
      </div>
      : <div className="accounts-table-wrap" aria-busy={listing.isFetching}>
        <table className="accounts-table">
          <caption className="sr-only">Catálogo de {plural}</caption>
          <thead><tr><th scope="col">Nombre</th><th scope="col">Estado</th><th scope="col">Acciones</th></tr></thead>
          <tbody>{listing.data.opciones.map(option => <tr key={option.id}>
            <th scope="row">{option.nombre}</th>
            <td><span className={option.activo ? 'account-active' : 'account-inactive'}>
              {option.activo ? (isArea ? 'Activa' : 'Activo') : (isArea ? 'Inactiva' : 'Inactivo')}
            </span></td>
            <td>
              <button type="button" className="account-action" aria-label={`Editar ${option.nombre}`}
                onClick={event => open({ action: 'edit', option }, event.currentTarget)}>Editar</button>
              {option.activo && <button type="button" className="account-action" aria-label={`Desactivar ${option.nombre}`}
                onClick={event => open({ action: 'deactivate', option }, event.currentTarget)}>Desactivar</button>}
            </td>
          </tr>)}</tbody>
        </table>
      </div>}
    {listing.data && !listing.isError && <div className="accounts-pagination">
      <p role="status">{total} {total === 1 ? 'resultado' : 'resultados'} · {page <= pages ? `Página ${page} de ${pages}` : `Página ${page} sin resultados`}</p>
      <div>
        <button type="button" className="shell-secondary" disabled={page <= 1 || listing.isFetching}
          onClick={() => updateFilters({ page: String(page - 1) })}>Anterior</button>
        <button type="button" className="shell-secondary" disabled={page >= pages || listing.isFetching}
          onClick={() => updateFilters({ page: String(page + 1) })}>Siguiente</button>
      </div>
    </div>}

    <dialog ref={dialog} className="shell-confirm classification-dialog" aria-labelledby="classification-editor-title"
      onCancel={event => {
        event.preventDefault();
        if (!mutation.isPending) closeEditor();
      }}
      onClose={() => {
        if (dialog.current?.open) return;
        if (opener.current?.isConnected) opener.current.focus();
        else createButton.current?.focus();
      }}>
      {editor && <form onSubmit={submit} aria-busy={mutation.isPending}>
        <h2 id="classification-editor-title">
          {editor.action === 'create' ? `Crear ${label}` :
            editor.action === 'edit' ? `Editar ${label}` : 'Confirmar desactivación'}
        </h2>
        {editor.action === 'deactivate' ? <p><strong>{editor.option.nombre}</strong> quedará disponible en las referencias históricas.</p>
          : <label>Nombre
            <input autoFocus value={name} readOnly={mutation.isPending} aria-invalid={Boolean(validation) || undefined}
              onChange={event => { setName(event.target.value); setValidation(''); mutation.reset(); }} />
          </label>}
        {validation && <p role="alert" className="access-error">{validation}</p>}
        {mutation.isError && <p role="alert" className="access-error">{errorMessage(mutation.error)}</p>}
        <div className="shell-dialog-actions">
          <button type="button" autoFocus={editor.action === 'deactivate'} className="shell-secondary"
            disabled={mutation.isPending} onClick={closeEditor}>Cancelar</button>
          <button type="submit" className="access-primary" disabled={mutation.isPending}>
            {mutation.isPending ? 'Guardando…' : editor.action === 'deactivate' ? 'Desactivar' : 'Guardar'}
          </button>
        </div>
      </form>}
    </dialog>
  </>;
}
