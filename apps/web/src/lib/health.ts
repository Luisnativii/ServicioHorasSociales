export type HealthRoute = 'live' | 'ready';

export async function health(path: HealthRoute, signal: AbortSignal): Promise<boolean> {
  try {
    const response = await fetch(`/api/health/${path}`, { signal, cache: 'no-store' });
    if (!response.ok) return false;
    const data: unknown = await response.json();
    return typeof data === 'object' && data !== null && 'status' in data && data.status === 'ok';
  } catch {
    return false;
  }
}
