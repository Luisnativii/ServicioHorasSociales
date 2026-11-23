import { useEffect, useRef, useState, type FormEvent } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import Hours from '../hours/Hours';
import Participations from '../participations/Participations';
import { Link, useSearchParams } from 'react-router-dom';
import { academicOptions, getStudent, listStudents, saveStudent, studentError, studentKey,
  type Student, type StudentDetail, type StudentInput } from './api';

type Options = Awaited<ReturnType<typeof academicOptions>>;
const emptyOptions: Options = { faculties: [], careers: [] };
const time = new Intl.DateTimeFormat('es-SV', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'America/El_Salvador' });
const pageSize = 25;

function StudentForm({ student, options, onSaved, onCancel }: {
  student?: StudentDetail; options: Options; onSaved: (student: Student) => void; onCancel: () => void;
}) {
  const [career, setCareer] = useState(student?.carrera.id ?? '');
  const current = options.careers.find(item => item.id === career);
  const faculty = options.faculties.find(item => item.id === current?.facultad_id);
  const changed = Boolean(student && career !== student.carrera.id);
  const save = useMutation({ mutationFn: (input: StudentInput) => saveStudent(input, student?.id), onSuccess: onSaved });

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (save.isPending) return;
    const form = new FormData(event.currentTarget);
    const value = (field: string) => String(form.get(field) ?? '').trim();
    save.mutate({
      carnet: value('carnet'), nombres: value('nombres') || null, apellidos: value('apellidos') || null,
      correo_institucional: value('correo'), carrera_id: career, motivo_cambio_carrera: value('motivo'),
    });
  }

  return <section className="institution-panel">
    <h2>{student ? 'Editar expediente' : 'Nuevo expediente'}</h2>
    {student && !student.nombres && <p>Nombre histórico: {student.nombre_completo}. Puedes conservarlo dejando nombre y apellido vacíos.</p>}
    <form className="institution-form" onSubmit={submit}>
      <label>Carnet <input name="carnet" inputMode="numeric" pattern="[0-9]{8}" maxLength={8} defaultValue={student?.carnet} required /></label>
      <label>Nombres <input name="nombres" maxLength={120} defaultValue={student?.nombres ?? ''} required={!student || !!student.nombres} /></label>
      <label>Apellidos <input name="apellidos" maxLength={120} defaultValue={student?.apellidos ?? ''} required={!student || !!student.apellidos} /></label>
      <label>Correo institucional <input name="correo" type="email" defaultValue={student?.correo_institucional} required /></label>
      <label>Carrera <select value={career} onChange={event => setCareer(event.target.value)} required
        disabled={Boolean(student && !student.historial_disponible)}>
        <option value="">Selecciona una carrera</option>
        {student && !options.careers.some(item => item.id === student.carrera.id) && <option value={student.carrera.id}>{student.carrera.nombre}</option>}
        {options.careers.filter(item => item.id === student?.carrera.id || (item.activo && item.facultad_activa))
          .map(item => <option key={item.id} value={item.id}>{item.nombre}</option>)}
      </select></label>
      <p>Facultad: {faculty?.nombre ?? (career === student?.carrera.id ? student?.facultad.nombre : 'Selecciona una carrera')}</p>
      {student && !student.historial_disponible && <p>El cambio de carrera estará disponible cuando el core habilite el historial académico.</p>}
      {changed && <label>Motivo del cambio <input name="motivo" maxLength={500} required /></label>}
      {save.error && <p role="alert" className="access-error">{studentError(save.error)}</p>}
      <div className="institution-actions">
        <button type="button" className="shell-secondary" onClick={onCancel} disabled={save.isPending}>Cancelar</button>
        <button type="submit" className="access-primary" disabled={save.isPending || !career}>{save.isPending ? 'Guardando…' : 'Guardar'}</button>
      </div>
    </form>
  </section>;
}

function Detail({ student, edit }: { student: StudentDetail; edit: () => void }) {
  return <section className="institution-panel">
    <h2>{student.nombre_completo}</h2>
    <dl className="institution-facts">
      <div><dt>Carnet</dt><dd>{student.carnet}</dd></div>
      <div><dt>Correo institucional</dt><dd>{student.correo_institucional}</dd></div>
      <div><dt>Carrera</dt><dd>{student.carrera.nombre}</dd></div>
      <div><dt>Facultad</dt><dd>{student.facultad.nombre}</dd></div>
      <div><dt>Estado</dt><dd>{student.activo ? 'Activo' : 'Inactivo'}</dd></div>
    </dl>
    <h3>Historial académico</h3>
    {!student.historial_disponible ? <p>El core todavía no ofrece el historial académico.</p>
      : student.vigencias.length === 0 ? <p>No hay cambios de carrera registrados en el historial disponible.</p>
        : <ul>{student.vigencias.map(period => <li key={period.desde ?? 'historica'}>
          {period.carrera.nombre} · {period.facultad.nombre} · {period.desde ? time.format(new Date(period.desde)) : 'Inicio desconocido'}
          {' — '}{period.hasta ? time.format(new Date(period.hasta)) : 'Vigente'}
          {period.motivo && <p>{period.motivo}</p>}
        </li>)}</ul>}
    <Participations key={student.id} view="estudiante" id={student.id} eligible={student.activo} />
    <Hours key={student.id} student={student.id} />
    <button type="button" className="shell-secondary" onClick={edit}>Editar expediente</button>
  </section>;
}

export default function Students() {
  const client = useQueryClient();
  const [params, setParams] = useSearchParams();
  const heading = useRef<HTMLHeadingElement>(null);
  const id = params.get('estudiante') ?? '';
  const view = params.get('vista');
  const creating = view === 'alta';
  const editing = id !== '' && view === 'editar';
  const showList = !id && !creating;
  const pageValue = Number(params.get('page') ?? 1);
  const page = Number.isSafeInteger(pageValue) && pageValue >= 1 && pageValue < 1e12 ? pageValue : 1;
  const search = (params.get('q') ?? '').slice(0, 100);
  const faculty = params.get('facultad_id') ?? '';
  const career = params.get('carrera_id') ?? '';
  const filters = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
  if (search) filters.set('q', search);
  if (faculty) filters.set('facultad_id', faculty);
  if (career) filters.set('carrera_id', career);

  const choices = useQuery({ queryKey: [...studentKey, 'catalogs'],
    queryFn: ({ signal }) => academicOptions(signal) });
  const options = choices.data ?? emptyOptions;
  const list = useQuery({ queryKey: [...studentKey, 'list', filters.toString()],
    queryFn: ({ signal }) => listStudents(filters, signal), enabled: showList });
  const detail = useQuery({ queryKey: [...studentKey, 'detail', id],
    queryFn: ({ signal }) => getStudent(id, signal), enabled: !!id && !creating });
  const pages = Math.max(1, Math.ceil((list.data?.pagina.total ?? 0) / pageSize));

  useEffect(() => { heading.current?.focus(); }, [id, view]);

  function next(changes: Record<string, string>) {
    const updated = new URLSearchParams(params);
    for (const [key, value] of Object.entries(changes)) {
      if (value) updated.set(key, value);
      else updated.delete(key);
    }
    return updated;
  }

  function saved(student: Student) {
    void client.invalidateQueries({ queryKey: studentKey });
    setParams(next({ estudiante: student.id, vista: '' }), { replace: true });
  }

  return <section className="accounts-page">
    <h1 ref={heading} tabIndex={-1}>Estudiantes</h1>
    {!showList && <Link className="institution-back shell-text-link"
      to={{ search: next({ estudiante: '', vista: '' }).toString() }}>Volver al listado</Link>}
    {choices.isError && <p role="alert">No se pudo cargar el catálogo académico.{' '}
      <button type="button" className="shell-secondary" onClick={() => void choices.refetch()}>Reintentar</button></p>}
    {creating && <StudentForm key="new" options={options} onSaved={saved} onCancel={() => setParams(next({ vista: '' }))} />}
    {!creating && id && <>
      {detail.isPending && <p role="status">Cargando expediente…</p>}
      {detail.isError && <p role="alert">{studentError(detail.error)}</p>}
      {!detail.isError && detail.data && (editing
        ? <StudentForm key={id} student={detail.data} options={options} onSaved={saved} onCancel={() => setParams(next({ vista: '' }))} />
        : <Detail student={detail.data} edit={() => setParams(next({ vista: 'editar' }))} />)}
    </>}
    {showList && <>
      <form className="accounts-toolbar academic-toolbar" onSubmit={event => {
        event.preventDefault();
        const query = String(new FormData(event.currentTarget).get('q') ?? '').trim();
        setParams(next({ q: query, page: '' }));
      }}>
        <label>Buscar por nombre o carnet <input type="search" name="q" maxLength={100} key={search} defaultValue={search} /></label>
        <button className="shell-secondary" type="submit">Buscar</button>
        <label>Facultad <select value={faculty} onChange={event => setParams(next({ facultad_id: event.target.value, carrera_id: '', page: '' }))}>
          <option value="">Todas</option>
          {options.faculties.map(item => <option key={item.id} value={item.id}>{item.nombre}</option>)}
        </select></label>
        <label>Carrera <select value={career} onChange={event => setParams(next({ carrera_id: event.target.value, page: '' }))}>
          <option value="">Todas</option>
          {options.careers.filter(item => !faculty || item.facultad_id === faculty)
            .map(item => <option key={item.id} value={item.id}>{item.nombre}</option>)}
        </select></label>
        <button className="shell-secondary" type="button" onClick={() => setParams(next({ vista: 'alta' }))}>Nuevo expediente</button>
      </form>
      {list.isPending && <p role="status">Cargando estudiantes…</p>}
      {list.isError && <p role="alert">{studentError(list.error)}</p>}
      {list.data && !list.isError && <>
        {list.data.estudiantes.length ? <div className="accounts-table-wrap"><table className="accounts-table academic-table" aria-label="Estudiantes">
          <thead><tr><th scope="col">Nombre</th><th scope="col">Carnet</th><th scope="col">Carrera</th><th scope="col">Facultad</th></tr></thead>
          <tbody>{list.data.estudiantes.map(student => <tr key={student.id}>
            <th scope="row"><Link to={{ search: next({ estudiante: student.id }).toString() }}>{student.nombre_completo}</Link></th>
            <td>{student.carnet}</td><td>{student.carrera.nombre}</td><td>{student.facultad.nombre}</td>
          </tr>)}</tbody>
        </table></div> : <p role="status">No hay expedientes en esta página.</p>}
        <nav className="accounts-pagination" aria-label="Páginas de estudiantes">
          <p aria-live="polite">{list.data.pagina.total} resultados · Página {page} de {pages}</p>
          <div>
            <button className="shell-secondary" type="button" disabled={page <= 1}
              onClick={() => setParams(next({ page: String(Math.min(page - 1, pages)) }))}>Anterior</button>
            <button className="shell-secondary" type="button" disabled={page >= pages}
              onClick={() => setParams(next({ page: String(page + 1) }))}>Siguiente</button>
          </div>
        </nav>
      </>}
    </>}
  </section>;
}
