import { api } from './api';

export type HealthRoute = 'live' | 'ready';

export async function health(path: HealthRoute, signal: AbortSignal): Promise<boolean> {
  try {
    const data: unknown = await api.request(`/api/health/${path}`, { signal, authenticated: false });
    return typeof data === 'object' && data !== null && 'status' in data && data.status === 'ok';
  } catch {
    return false;
  }
}
