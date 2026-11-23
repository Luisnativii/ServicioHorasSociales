import { useState, type FormEvent } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { api, ApiError } from '../../lib/api';
import { listStudents } from '../students/api';

type Participant = {
  id: string;
  proyecto_id: string;
  estudiante_id: string;
  proyecto: string;
  estudiante: string;
  carnet: string;
  estado: string;
  asignada_en: string;
  cerrada_en: string | null;
};
type Candidate = { id: string; label: string; eligible: boolean };
type Pagination = { total: number };
type View = 'proyecto' | 'estudiante';

const pageSize = 10;
const dates = new Intl.DateTimeFormat('es-SV', {
  dateStyle: 'medium', timeStyle: 'short', timeZone: 'America/El_Salvador',
});
const labels: Record<string, string> = {
  ASIGNADA: 'Asignada', EN_CURSO: 'En curso', FINALIZADA: 'Finalizada',
  RETIRADA: 'Retirada', CANCELADA: 'Cancelada',
};

function participationError(error: Error) {
  if (error instanceof ApiError) {
    switch (error.field) {
      case 'estado_proyecto': return 'El proyecto ya no está activo. No se guardó la asignación.';
      case 'estado_estudiante': return 'El estudiante está inactivo.';
      case 'participacion': return 'Este estudiante ya tiene una participación en el proyecto.';
      case 'escritura_core': return 'Las asignaciones todavía no están habilitadas en este entorno.';
    }
    if (error.status === 404) return 'El proyecto o expediente ya no está disponible.';
  }
  return 'No se pudo completar la operación. Consulta las participaciones antes de volver a intentarlo.';
}

async function candidates(view: View, search: string, page: number, signal: AbortSignal) {
  const filters = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
  if (search) filters.set('q', search);
  if (view === 'proyecto') {
    const result = await listStudents(filters, signal);
    return {
      items: result.estudiantes.map(student => ({
        id: student.id, label: `${student.nombre_completo} · ${student.carnet}`, eligible: student.activo,
      })), total: result.pagina.total,
    };
  }
  filters.set('estado', 'ACTIVO');
  const result = await api.request<{
    proyectos: { id: string; titulo: string; institucion: string }[]; pagina: Pagination;
  }>(`/api/v1/proyectos?${filters}`, { signal });
  return {
    items: result.proyectos.map(project => ({
      id: project.id, label: `${project.titulo} · ${project.institucion}`, eligible: true,
    })), total: result.pagina.total,
  };
}

function Pager({ page, total, change, label }: { page: number; total: number; change: (page: number) => void; label: string }) {
  const pages = Math.max(1, Math.ceil(total / pageSize));
  return <nav className="accounts-pagination" aria-label={label}>
    <p>{total} resultados · Página {page} de {pages}</p>
    <div><button type="button" className="shell-secondary" disabled={page <= 1} onClick={() => change(page - 1)}>Anterior</button>
    <button type="button" className="shell-secondary" disabled={page >= pages} onClick={() => change(page + 1)}>Siguiente</button></div>
  </nav>;
}

function Assignment({ view, id, saved }: { view: View; id: string; saved: () => void }) {
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const [selected, setSelected] = useState<Candidate>();
  const query = useQuery({
    queryKey: ['participation-candidates', view, search, page],
    queryFn: ({ signal }) => candidates(view, search, page, signal),
  });
  const assign = useMutation({
    mutationFn: ({ candidate, confirmed }: { candidate: Candidate; confirmed: boolean }) => {
      const project = view === 'proyecto' ? id : candidate.id;
      return api.request<{ id: string }>(`/api/v1/proyectos/${encodeURIComponent(project)}/participaciones`, {
        method: 'POST', body: {
          estudiante_id: view === 'proyecto' ? candidate.id : id, confirmar_exceso_cupo: confirmed,
        },
      });
    },
    onSuccess: saved,
  });
  const warning = assign.error instanceof ApiError && assign.error.field === 'cupo';
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const term = String(new FormData(event.currentTarget).get('q') ?? '').trim();
    setSearch(term);
    setPage(1);
    setSelected(undefined);
    assign.reset();
  }

  return <div>
    <form className="accounts-toolbar academic-toolbar" onSubmit={submit}>
      <label>{view === 'proyecto' ? 'Buscar por nombre o carnet' : 'Buscar proyectos activos'}
        <input name="q" type="search" maxLength={view === 'proyecto' ? 100 : 200} disabled={assign.isPending} />
      </label>
      <button type="submit" className="shell-secondary" disabled={assign.isPending}>Buscar</button>
    </form>
    {query.isPending && <p role="status">Buscando…</p>}
    {query.isError && <p role="alert">{participationError(query.error)}</p>}
    {query.data && !query.isError && <>
      {query.data.items.length ? <ul>{query.data.items.map(candidate => <li key={candidate.id}>
        <button type="button" className="shell-secondary" aria-pressed={selected?.id === candidate.id}
          disabled={!candidate.eligible || assign.isPending} onClick={() => { setSelected(candidate); assign.reset(); }}>
          {candidate.label}{!candidate.eligible && ' · Inactivo'}
        </button>
      </li>)}</ul> : <p>No hay candidatos en esta página.</p>}
      {!assign.isPending && <Pager page={page} total={query.data.total} change={value => {
        setPage(value); setSelected(undefined); assign.reset();
      }} label="Páginas de candidatos" />}
    </>}
    {selected && <p>Selección: {selected.label}</p>}
    {assign.error && !warning && <p role="alert">{participationError(assign.error)}</p>}
    {warning && <p role="alert">El cupo solicitado está completo. Confirma si quieres asignar de todos modos.</p>}
    <button type="button" className="access-primary" disabled={!selected || assign.isPending} onClick={() => {
      if (selected) assign.mutate({ candidate: selected, confirmed: warning });
    }}>{assign.isPending ? 'Guardando…' : warning ? 'Confirmar exceso de cupo' : 'Asignar'}</button>
  </div>;
}

export default function Participations({ view, id, eligible }: { view: View; id: string; eligible: boolean }) {
  const client = useQueryClient();
  const [page, setPage] = useState(1);
  const [adding, setAdding] = useState(false);
  const [saved, setSaved] = useState(false);
  const collection = view === 'proyecto' ? 'proyectos' : 'estudiantes';
  const query = useQuery({
    queryKey: ['participations', view, id, page],
    queryFn: ({ signal }) => api.request<{ participaciones: Participant[]; pagina: Pagination }>(
      `/api/v1/${collection}/${encodeURIComponent(id)}/participaciones?page=${page}&page_size=${pageSize}`, { signal }),
  });
  return <section aria-label="Participaciones">
    <h3>Participaciones</h3>
    {saved && <p role="status">Asignación guardada.</p>}
    {eligible && <button type="button" className="shell-secondary" aria-expanded={adding} onClick={() => {
      setAdding(!adding); setSaved(false);
    }}>{adding ? 'Cerrar asignación' : view === 'proyecto' ? 'Asignar estudiante' : 'Asignar a proyecto'}</button>}
    {adding && eligible && <Assignment id={id} view={view} saved={() => {
      setAdding(false); setSaved(true); setPage(1);
      void client.invalidateQueries({ queryKey: ['participations'] });
    }} />}
    {query.isPending && <p role="status">Cargando participaciones…</p>}
    {query.isError && <p role="alert">{participationError(query.error)}</p>}
    {query.data && !query.isError && <>
      {query.data.participaciones.length ? <div className="accounts-table-wrap"><table className="accounts-table">
        <thead><tr><th scope="col">{view === 'proyecto' ? 'Estudiante' : 'Proyecto'}</th><th scope="col">Estado</th>
          <th scope="col">Asignación</th><th scope="col">Cierre</th></tr></thead>
        <tbody>{query.data.participaciones.map(row => <tr key={row.id}>
          <th scope="row">{view === 'proyecto'
            ? <Link to={`/estudiantes?estudiante=${row.estudiante_id}`}>{row.estudiante} · {row.carnet}</Link>
            : <Link to={`/proyectos?proyecto=${row.proyecto_id}`}>{row.proyecto}</Link>}</th>
          <td>{labels[row.estado] ?? row.estado}</td><td>{dates.format(new Date(row.asignada_en))}</td>
          <td>{row.cerrada_en ? dates.format(new Date(row.cerrada_en)) : 'Sin cierre'}</td>
        </tr>)}</tbody>
      </table></div> : <p>Sin participaciones en esta página.</p>}
      <Pager page={page} total={query.data.pagina.total} change={setPage} label="Páginas de participaciones" />
    </>}
  </section>;
}
