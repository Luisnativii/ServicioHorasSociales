import { createContext, useContext, useEffect, useSyncExternalStore, type ReactNode } from 'react';
import { QueryClient } from '@tanstack/react-query';
import { ApiClient, ApiError, api } from '../lib/api';

export type Administrator = { id: string; correo: string };
export type SessionState =
  | { status: 'loading' | 'anonymous'; user: null; error: null }
  | { status: 'authenticated'; user: Administrator; error: null }
  | { status: 'error'; user: null; error: Error };

export function createQueryClient(): QueryClient {
  return new QueryClient({ defaultOptions: {
    queries: { retry: false, refetchOnWindowFocus: false },
    mutations: { retry: false },
  } });
}

function identity(value: Administrator): Administrator {
  if (!value || typeof value.id !== 'string' || typeof value.correo !== 'string') throw new Error('Respuesta de identidad no válida.');
  return { id: value.id, correo: value.correo };
}

export class SessionController {
  private state: SessionState = { status: 'loading', user: null, error: null };
  private listeners = new Set<() => void>();
  private version = 0;
  private pending?: Promise<void>;
  private busy = false;

  constructor(readonly queries: QueryClient, private readonly client: ApiClient) {
    client.onUnauthorized(() => this.reset());
  }

  getSnapshot = (): SessionState => this.state;
  subscribe = (listener: () => void): (() => void) => {
    this.listeners.add(listener);
    return () => { this.listeners.delete(listener); };
  };
  private set(state: SessionState): void { this.state = state; this.listeners.forEach(listener => listener()); }

  private clear(): void {
    this.version++;
    this.pending = undefined;
    this.client.invalidate();
    void this.queries.cancelQueries();
    this.queries.clear();
  }
  private reset(): void { this.clear(); this.set({ status: 'anonymous', user: null, error: null }); }

  validate = (): Promise<void> => {
    if (this.busy) return Promise.resolve();
    if (this.pending) return this.pending;
    const version = this.version;
    const pending = (async () => {
      try {
        const user = identity(await this.queries.fetchQuery({ queryKey: ['session'], staleTime: 0,
          queryFn: ({ signal }) => this.client.request<Administrator>('/api/v1/auth/me', { signal }),
        }));
        if (version === this.version) this.set({ status: 'authenticated', user, error: null });
      } catch (error) {
        if (version !== this.version) return;
        if (error instanceof ApiError && error.status === 401) this.reset();
        else {
          this.clear();
          this.set({ status: 'error', user: null, error: error instanceof Error ? error : new Error('No se pudo comprobar la sesión.') });
        }
      }
    })();
    this.pending = pending;
    void pending.finally(() => { if (this.pending === pending) this.pending = undefined; });
    return pending;
  };

  async login(correo: string, contrasena: string): Promise<void> {
    if (this.busy) throw new Error('Hay una operación de acceso en curso.');
    this.busy = true;
    this.clear();
    this.set({ status: 'loading', user: null, error: null });
    try {
      await this.client.prepareCSRF();
      const user = identity(await this.client.request<Administrator>('/api/v1/auth/login', {
        method: 'POST', body: { correo, contrasena }, authenticated: false,
      }));
      this.queries.setQueryData(['session'], user);
      this.set({ status: 'authenticated', user, error: null });
    } catch (error) {
      this.reset();
      throw error;
    } finally { this.busy = false; }
  }

  async logout(): Promise<void> {
    if (this.busy) throw new Error('Hay una operación de acceso en curso.');
    this.busy = true;
    // Hide cached private data immediately, even when the network result is uncertain.
    this.reset();
    try {
      await this.client.prepareCSRF();
      await this.client.request<void>('/api/v1/auth/logout', { method: 'POST', authenticated: false });
    } finally { this.reset(); this.busy = false; }
  }
}

export const queryClient = createQueryClient();
export const session = new SessionController(queryClient, api);
const Context = createContext<SessionController | null>(null);

export function SessionProvider({ children }: { children: ReactNode }) {
  useEffect(() => {
    void session.validate();
    const check = () => { if (document.visibilityState === 'visible') void session.validate(); };
    window.addEventListener('focus', check);
    document.addEventListener('visibilitychange', check);
    const timer = window.setInterval(check, 60_000);
    return () => { window.removeEventListener('focus', check); document.removeEventListener('visibilitychange', check); window.clearInterval(timer); };
  }, []);
  return <Context.Provider value={session}>{children}</Context.Provider>;
}

export function useSession() {
  const controller = useContext(Context);
  if (!controller) throw new Error('Falta SessionProvider.');
  const state = useSyncExternalStore(controller.subscribe, controller.getSnapshot);
  return { ...state, login: controller.login.bind(controller), logout: controller.logout.bind(controller), validate: controller.validate };
}
