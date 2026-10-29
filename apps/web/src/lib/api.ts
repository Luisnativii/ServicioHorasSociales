export class ApiError extends Error {
  constructor(public readonly status: number, public readonly code: string, public readonly retryAfter?: number) {
    super(status === 401 ? 'Debes iniciar sesión.' : status === 429 ? 'Espera antes de volver a intentarlo.' : 'No se pudo completar la solicitud.');
  }
}
export class NetworkError extends Error {
  constructor() { super('No se pudo conectar con el servidor.'); }
}
export class StaleSessionError extends Error {
  constructor() { super('La sesión cambió durante la solicitud.'); }
}

type RequestOptions = {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  body?: unknown;
  signal?: AbortSignal;
  authenticated?: boolean;
};

export function readCSRF(): string | undefined {
  const matches = document.cookie.split(';').map(c => c.trim()).filter(c => c.startsWith('hs_csrf='));
  if (matches.length !== 1) return undefined;
  try { return decodeURIComponent(matches[0].slice('hs_csrf='.length)); } catch { return undefined; }
}

// A single transport for queries and mutations. No refresh or automatic retries.
export class ApiClient {
  private generation = 0;
  private csrfPending?: Promise<void>;
  private csrfGeneration = 0;
  private unauthorized = new Set<() => void>();

  constructor(private readonly send: typeof fetch = (...args) => fetch(...args), private readonly csrfCookie = readCSRF) {}

  onUnauthorized(callback: () => void): () => void {
    this.unauthorized.add(callback);
    return () => { this.unauthorized.delete(callback); };
  }

  invalidate(): void { this.generation++; }

  private path(path: string): string {
    const base = 'https://portal.invalid';
    const parsed = new URL(path, base);
    if (!path.startsWith('/api/') || parsed.origin !== base || !parsed.pathname.startsWith('/api/') || parsed.hash) {
      throw new Error('La solicitud debe usar una ruta relativa /api/.');
    }
    return parsed.pathname + parsed.search;
  }

  private async response(path: string, init: RequestInit): Promise<Response> {
    try { return await this.send(this.path(path), { ...init, credentials: 'same-origin', cache: 'no-store', redirect: 'error' }); }
    catch (error) {
      if (init.signal?.aborted) throw error;
      throw new NetworkError();
    }
  }

  private async failure(response: Response): Promise<ApiError> {
    let code = 'http_error';
    try {
      const data = await response.json();
      if (typeof data?.error?.code === 'string') code = data.error.code;
    } catch { /* Error pages are not JSON; never display their raw body. */ }
    const seconds = Number(response.headers.get('Retry-After'));
    return new ApiError(response.status, code, response.status === 429 && Number.isFinite(seconds) && seconds > 0 ? seconds : undefined);
  }

  prepareCSRF(): Promise<void> {
    if (this.csrfPending) {
      if (this.csrfGeneration === this.generation) return this.csrfPending;
      // Serialize cookie-setting requests across identity changes as well.
      return this.csrfPending.catch(() => {}).then(() => this.prepareCSRF());
    }
    const generation = this.generation;
    this.csrfGeneration = generation;
    const pending = (async () => {
      const response = await this.response('/api/v1/auth/csrf', { method: 'GET' });
      if (generation !== this.generation) throw new StaleSessionError();
      if (!response.ok) throw await this.failure(response);
      if (!this.csrfCookie()) throw new ApiError(403, 'csrf_unavailable');
    })();
    this.csrfPending = pending;
    void pending.finally(() => { if (this.csrfPending === pending) this.csrfPending = undefined; }).catch(() => {});
    return pending;
  }

  async request<T>(path: string, options: RequestOptions = {}): Promise<T> {
    const generation = this.generation;
    // Public service probes have no private data to discard on a session change.
    const sessionScoped = path !== '/api/health/live' && path !== '/api/health/ready';
    const method = options.method ?? 'GET';
    const headers = new Headers({ Accept: 'application/json' });
    if (method !== 'GET') {
      if (!this.csrfCookie()) await this.prepareCSRF();
      if (generation !== this.generation) throw new StaleSessionError();
      headers.set('X-CSRF-Token', this.csrfCookie()!);
    }
    if (options.body !== undefined) headers.set('Content-Type', 'application/json');
    const response = await this.response(path, {
      method, headers, signal: options.signal,
      body: options.body === undefined ? undefined : JSON.stringify(options.body),
    });
    if (sessionScoped && generation !== this.generation) throw new StaleSessionError();
    if (!response.ok) {
      const error = await this.failure(response);
      if (sessionScoped && generation !== this.generation) throw new StaleSessionError();
      if (response.status === 401 && options.authenticated !== false) {
        this.invalidate();
        this.unauthorized.forEach(callback => callback());
      }
      throw error;
    }
    const result = response.status === 204 ? undefined : await response.json();
    if (sessionScoped && generation !== this.generation) throw new StaleSessionError();
    return result as T;
  }
}

export const api = new ApiClient();
