import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { api } from '../../lib/api';

type Ledger = {
  movimientos: {
    id: string; proyecto_id: string; proyecto: string; horas: number; naturaleza: 'INTERNA' | 'EXTERNA';
    fecha_horas: string; registrado_en: string; origen: 'IMPORTACION' | 'MANUAL'; concepto: string | null;
    administrador_id: string | null; usuario_core_id: string | null;
  }[];
  avance: { internas: number; externas: number; internas_pendientes: number; externas_pendientes: number; meta_por_categoria: number };
  pagina: { total: number };
};
const size = 10;
const registrationDate = new Intl.DateTimeFormat('es-SV', {
  dateStyle: 'medium', timeStyle: 'short', timeZone: 'America/El_Salvador',
});

export default function Hours({ student }: { student: string }) {
  const [page, setPage] = useState(1);
  const query = useQuery({
    queryKey: ['hours', student, page],
    queryFn: ({ signal }) => api.request<Ledger>(
      `/api/v1/estudiantes/${encodeURIComponent(student)}/horas?page=${page}&page_size=${size}`, { signal }),
  });
  const data = query.data;
  const pages = Math.max(1, Math.ceil((data?.pagina.total ?? 0) / size));

  return <section aria-label="Libro de horas">
    <h3>Horas</h3>
    {query.isPending && <p role="status">Consultando el libro de horas…</p>}
    {query.isError && <div role="alert">No se pudo consultar el libro de horas.{' '}
      <button type="button" className="shell-secondary" onClick={() => void query.refetch()}>Reintentar</button></div>}
    {data && !query.isError && <>
      <dl className="institution-facts">
        <div><dt>Internas</dt><dd>{data.avance.internas} de {data.avance.meta_por_categoria} · {data.avance.internas_pendientes} pendientes</dd></div>
        <div><dt>Externas</dt><dd>{data.avance.externas} de {data.avance.meta_por_categoria} · {data.avance.externas_pendientes} pendientes</dd></div>
      </dl>
      {data.movimientos.length ? <div className="accounts-table-wrap"><table className="accounts-table">
        <thead><tr><th scope="col">Proyecto y concepto</th><th scope="col">Horas</th><th scope="col">Actividad</th>
          <th scope="col">Registro</th><th scope="col">Origen</th></tr></thead>
        <tbody>{data.movimientos.map(row => <tr key={row.id}>
          <th scope="row"><Link to={`/proyectos?proyecto=${row.proyecto_id}`}>{row.proyecto}</Link>{row.concepto && <p>{row.concepto}</p>}</th>
          <td>{row.horas} {row.naturaleza === 'INTERNA' ? 'internas' : 'externas'}</td>
          <td>{row.fecha_horas}</td><td>{registrationDate.format(new Date(row.registrado_en))}</td>
          <td>{row.origen === 'IMPORTACION' ? 'Importación' : 'Manual'}
            {row.administrador_id && <p>ADMIN del portal: {row.administrador_id}</p>}
            {row.usuario_core_id && <p>Usuario del core: {row.usuario_core_id}</p>}
          </td>
        </tr>)}</tbody>
      </table></div> : <p>Sin movimientos en esta página.</p>}
      {pages > 1 && <nav className="accounts-pagination" aria-label="Páginas del libro de horas">
        <p>{data.pagina.total} movimientos · Página {page} de {pages}</p>
        <div>
          <button type="button" className="shell-secondary" disabled={page <= 1} onClick={() => setPage(page - 1)}>Anterior</button>
          <button type="button" className="shell-secondary" disabled={page >= pages} onClick={() => setPage(page + 1)}>Siguiente</button>
        </div>
      </nav>}
    </>}
  </section>;
}
