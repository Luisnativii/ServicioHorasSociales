import { api, ApiClient, ApiError, NetworkError } from '../../lib/api';

export type Account = { id: string; correo: string; activo: boolean };
export type AccountsPage = { cuentas: Account[]; pagina: { pagina: number; page_size: number; total: number } };

function account(value: unknown): Account {
  const v = value as Account;
  if (!v || typeof v.id !== 'string' || !/^[a-f0-9]{8}(-[a-f0-9]{4}){3}-[a-f0-9]{12}$/.test(v.id)
    || typeof v.correo !== 'string' || typeof v.activo !== 'boolean') throw new Error('Respuesta de cuenta no válida.');
  return { id: v.id, correo: v.correo, activo: v.activo };
}

export function accountsApi(client: ApiClient = api) {
  return {
    async list(page: number, size: number, order: 'asc' | 'desc', signal?: AbortSignal): Promise<AccountsPage> {
      if (!Number.isSafeInteger(page) || page < 1 || ![10, 25, 50].includes(size)) throw new Error('Paginación inválida.');
      const query = new URLSearchParams({ page: String(page), page_size: String(size), sort: 'creado_en', order });
      const value = await client.request<AccountsPage>(`/api/v1/admin/cuentas?${query}`, { signal });
      if (!value || !Array.isArray(value.cuentas) || value.cuentas.length > size || value.pagina?.pagina !== page
        || value.pagina.page_size !== size || !Number.isSafeInteger(value.pagina.total) || value.pagina.total < 0) throw new Error('Respuesta de listado no válida.');
      return { cuentas: value.cuentas.map(account), pagina: { pagina: page, page_size: size, total: value.pagina.total } };
    },
    async setActive(selected: Account): Promise<Account> {
      const previous = account(selected);
      const action = previous.activo ? 'desactivar' : 'activar';
      const result = account(await client.request(`/api/v1/admin/cuentas/${previous.id}/${action}`, { method: 'POST' }));
      if (result.id !== previous.id || result.activo === previous.activo) throw new Error('No se pudo confirmar el estado de la cuenta.');
      return result;
    },
  };
}

export function accountError(error: unknown): string {
  if (error instanceof NetworkError) return 'No pudimos conectar con el portal.';
  if (error instanceof ApiError) {
    if (error.status === 403) return 'No se pudo validar el cambio. Actualiza la página antes de volver a intentarlo.';
    if (error.status === 404) return 'La cuenta ya no está disponible. Actualiza el listado.';
    if (error.status === 401) return 'Tu sesión ha vencido. Vuelve a iniciar sesión.';
    if (error.status === 409) return 'El estado de la cuenta cambió. Actualiza el listado.';
  }
  return 'El portal no pudo completar la solicitud.';
}

export function uncertainChange(error: unknown): boolean {
  return !(error instanceof ApiError && [400, 401, 403, 404, 409].includes(error.status));
}
