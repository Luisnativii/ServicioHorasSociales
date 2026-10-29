import { useEffect, useRef, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSearchParams } from 'react-router-dom';
import { useSession } from '../../app/session';
import { accountsApi, accountError, uncertainChange, type Account, type AccountsPage } from './api';

const accounts = accountsApi();

export default function Accounts() {
  const session = useSession();
  const client = useQueryClient();
  const [params, setParams] = useSearchParams();
  const pageInput = Number(params.get('page') ?? '1');
  const page = Number.isSafeInteger(pageInput) && pageInput > 0 ? pageInput : 1;
  const sizeInput = Number(params.get('page_size') ?? '25');
  const size = [10, 25, 50].includes(sizeInput) ? sizeInput : 25;
  const order = params.get('order') === 'desc' ? 'desc' : 'asc';
  const query = useQuery({ queryKey: ['accounts', page, size, order], queryFn: ({ signal }) => accounts.list(page, size, order, signal) });
  const dialog = useRef<HTMLDialogElement>(null);
  const opener = useRef<HTMLButtonElement | null>(null);
  const heading = useRef<HTMLHeadingElement>(null);
  const submitting = useRef(false);
  const [selected, setSelected] = useState<Account | null>(null);
  const [notice, setNotice] = useState('');
  const [error, setError] = useState('');
  const [uncertain, setUncertain] = useState(false);
  const [reloading, setReloading] = useState(false);

  const mutation = useMutation({
    mutationFn: accounts.setActive,
    onSuccess: async result => {
      // Only update after the server confirms the change, never optimistically.
      client.setQueriesData<AccountsPage>({ queryKey: ['accounts'] }, data => data ? {
        ...data, cuentas: data.cuentas.map(row => row.id === result.id ? result : row),
      } : data);
      setNotice(`Cuenta ${result.correo} ${result.activo ? 'activada' : 'desactivada'}.`);
      await client.invalidateQueries({ queryKey: ['accounts'] });
      dialog.current?.close();
    },
    onError: failure => { setError(accountError(failure)); setUncertain(uncertainChange(failure)); },
    onSettled: () => { submitting.current = false; },
  });

  useEffect(() => { if (selected && !dialog.current?.open) dialog.current?.showModal(); }, [selected]);

  function choose(row: Account, button: HTMLButtonElement) {
    opener.current = button; setError(''); setUncertain(false); setNotice(''); mutation.reset(); setSelected(row);
  }
  function paginate(next: number, nextSize = size, nextOrder = order) {
    setNotice(''); setParams({ page: String(next), page_size: String(nextSize), order: nextOrder });
  }
  async function reconcile() {
    setReloading(true);
    try { await query.refetch({ throwOnError: true }); dialog.current?.close(); setNotice('Listado actualizado. Revisa el estado de la cuenta antes de decidir otro cambio.'); }
    catch (failure) { setError(accountError(failure)); }
    finally { setReloading(false); }
  }

  const total = query.data?.pagina.total;
  const pages = total === undefined ? 1 : Math.max(1, Math.ceil(total / size));
  const busy = mutation.isPending || reloading;

  return <section className="accounts-page">
    <p className="access-eyebrow">Administración</p>
    <h1 ref={heading} tabIndex={-1}>Cuentas administradoras</h1>
    <p className="shell-lead">Consulta quién puede acceder al portal y administra el estado de sus cuentas.</p>
    <div className="accounts-toolbar">
      <label>Orden de creación<select value={order} onChange={e => paginate(1, size, e.target.value as 'asc' | 'desc')}>
        <option value="asc">Más antiguas primero</option><option value="desc">Más recientes primero</option>
      </select></label>
      <label>Cuentas por página<select value={size} onChange={e => paginate(1, Number(e.target.value))}>
        {[10, 25, 50].map(n => <option key={n} value={n}>{n}</option>)}
      </select></label>
      <button className="shell-secondary" disabled={query.isFetching} onClick={() => void query.refetch()}>Actualizar listado</button>
    </div>
    {notice && <p role="status" className="accounts-success">{notice}</p>}
    {query.isPending ? <p role="status" className="accounts-feedback">Cargando cuentas…</p>
      : query.isError ? <div role="alert" className="accounts-feedback"><p>{accountError(query.error)} Vuelve a actualizar el listado.</p></div>
      : query.data.cuentas.length === 0 ? <div className="shell-empty"><h2>No hay cuentas en esta página</h2><p>Actualiza el listado o vuelve a la primera página.</p>
        {page > 1 && <button className="shell-secondary" onClick={() => paginate(1)}>Primera página</button>}</div>
      : <div className="accounts-table-wrap" aria-busy={query.isFetching}>
        <table className="accounts-table"><caption className="sr-only">Cuentas administradoras y estado de acceso</caption>
          <thead><tr><th scope="col">Correo electrónico</th><th scope="col">Estado</th><th scope="col">Acción</th></tr></thead>
          <tbody>{query.data.cuentas.map(row => <tr key={row.id}>
            <th scope="row">{row.correo}{row.id === session.user?.id && <small>Tu cuenta</small>}</th>
            <td><span className={row.activo ? 'account-active' : 'account-inactive'}>{row.activo ? 'Activa' : 'Inactiva'}</span></td>
            <td><button className="account-action" disabled={busy} aria-label={`${row.activo ? 'Desactivar' : 'Activar'} cuenta ${row.correo}`}
              onClick={e => choose(row, e.currentTarget)}>{row.activo ? 'Desactivar' : 'Activar'}</button></td>
          </tr>)}</tbody>
        </table>
      </div>}
    {total !== undefined && !query.isError && <div className="accounts-pagination">
      <p>{total} {total === 1 ? 'cuenta' : 'cuentas'} · Página {page} de {pages}</p>
      <div><button className="shell-secondary" disabled={page <= 1 || query.isFetching} onClick={() => paginate(page - 1)}>Anterior</button>
        <button className="shell-secondary" disabled={page >= pages || query.isFetching} onClick={() => paginate(page + 1)}>Siguiente</button></div>
    </div>}
    <aside className="accounts-policy" aria-label="Política de acceso">
      <h2>Sobre las cuentas</h2><p>Las cuentas nuevas se crean mediante la herramienta de administración del sistema.</p>
      <p>Es posible desactivar tu propia cuenta o dejar todas inactivas. Si esto ocurre, la persona responsable del sistema puede crear otra cuenta con esa herramienta.</p>
    </aside>

    <dialog ref={dialog} className="shell-confirm" aria-labelledby="account-change-title" aria-describedby="account-change-description"
      onCancel={e => { if (busy) e.preventDefault(); }} onClose={() => {
        setSelected(null); if (opener.current?.isConnected) opener.current.focus(); else heading.current?.focus();
      }}>
      {selected && <>
        <h2 id="account-change-title">¿{selected.activo ? 'Desactivar' : 'Activar'} esta cuenta?</h2>
        <p className="account-dialog-email">{selected.correo}</p>
        <p id="account-change-description">{selected.activo
          ? 'No podrá iniciar nuevas sesiones. Las sesiones que ya estén abiertas continuarán hasta su vencimiento.'
          : 'Esta cuenta podrá volver a iniciar sesión con su contraseña actual.'}</p>
        {selected.id === session.user?.id && selected.activo && <p className="account-self-warning">Esta es tu cuenta. Podrás continuar en esta sesión hasta que venza, pero no podrás volver a entrar mientras esté inactiva.</p>}
        {error && <p role="alert" className="access-error">{error}{uncertain && ' No sabemos si el cambio se completó. Actualiza el listado antes de intentarlo de nuevo.'}</p>}
        <div className="shell-dialog-actions">
          <button className="shell-secondary" autoFocus disabled={busy} onClick={() => dialog.current?.close()}>Cancelar</button>
          {uncertain ? <button className="access-primary" disabled={busy} onClick={() => void reconcile()}>{reloading ? 'Actualizando…' : 'Actualizar listado'}</button>
            : <button className="access-primary" disabled={busy} onClick={() => {
              if (submitting.current) return; submitting.current = true; setError(''); mutation.mutate(selected);
            }}>{mutation.isPending ? 'Guardando…' : selected.activo ? 'Desactivar cuenta' : 'Activar cuenta'}</button>}
        </div>
      </>}
    </dialog>
  </section>;
}
