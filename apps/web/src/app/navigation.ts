// Add only implemented, protected destinations here; never infer a return URL.
export const navigation = [
  { path: '/', label: 'Inicio', title: 'Inicio' },
  { path: '/administradores', label: 'Administradores', title: 'Cuentas administradoras' },
  { path: '/sistema/estado', label: 'Estado del portal', title: 'Estado del portal' },
] as const;

export function authorizedDestination(value: unknown): string {
  return typeof value === 'string' && navigation.some(item => item.path === value) ? value : '/';
}
