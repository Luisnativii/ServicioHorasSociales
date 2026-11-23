// Add only implemented, protected destinations here; never infer a return URL.
export const navigation = [
  { path: '/', label: 'Inicio', title: 'Inicio' },
  { path: '/instituciones', label: 'Instituciones', title: 'Instituciones' },
  { path: '/proyectos', label: 'Proyectos', title: 'Proyectos' },
  { path: '/estudiantes', label: 'Estudiantes', title: 'Estudiantes' },
  { path: '/catalogos/academico', label: 'Catálogo académico', title: 'Catálogo académico' },
  { path: '/catalogos/territorial', label: 'Catálogo territorial', title: 'Catálogo territorial' },
  { path: '/catalogos/clasificacion', label: 'Clasificación de proyectos', title: 'Clasificación de proyectos' },
  { path: '/administradores', label: 'Administradores', title: 'Cuentas administradoras' },
  { path: '/sistema/estado', label: 'Estado del portal', title: 'Estado del portal' },
] as const;

export function authorizedDestination(value: unknown): string {
  return typeof value === 'string' && navigation.some(item => item.path === value) ? value : '/';
}
