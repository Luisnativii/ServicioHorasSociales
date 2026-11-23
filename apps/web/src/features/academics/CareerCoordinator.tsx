import { useEffect, useId, useRef, useState, type FormEvent } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ApiError, NetworkError } from '../../lib/api';
import { getCareerCoordinator, saveCareerCoordinator, type Career } from './api';

function saveError(error: Error) {
  if (error instanceof NetworkError) return 'Se perdió la conexión durante el guardado. Tu edición sigue en el formulario.';
  if (error instanceof ApiError && error.status === 404) return 'La carrera ya no está disponible.';
  if (error instanceof ApiError && error.status === 400) return 'Revisa el nombre e inténtalo de nuevo.';
  return 'No se pudo guardar el nombre. Tu edición sigue en el formulario.';
}

export default function CareerCoordinator({ career }: { career: Career }) {
  const fieldId = useId();
  const queryClient = useQueryClient();
  const queryKey = ['academic-catalog', 'coordinator', career.id];
  const editable = career.activo && career.facultad_activa;
  const [draft, setDraft] = useState<string | null>(null);
  const [invalid, setInvalid] = useState(false);
  const [notice, setNotice] = useState('');
  const submitting = useRef(false);
  const editButton = useRef<HTMLButtonElement>(null);
  const restoreFocus = useRef(false);

  const coordinator = useQuery({
    queryKey,
    queryFn: ({ signal }) => getCareerCoordinator(career.id, signal),
  });

  const currentName = coordinator.data?.coordinador_nombre ?? null;

  const save = useMutation({
    mutationFn: (name: string | null) => saveCareerCoordinator(career.id, name),
    onMutate: () => queryClient.cancelQueries({ queryKey, exact: true }),
    onSuccess: result => {
      queryClient.setQueryData(queryKey, current => current === undefined ? undefined : result);
      restoreFocus.current = true;
      setNotice(result.coordinador_nombre === null ? 'Nombre retirado.' : 'Nombre guardado.');
      setDraft(null);
    },
    onSettled: () => { submitting.current = false; },
  });

  useEffect(() => {
    if (draft === null && restoreFocus.current) {
      editButton.current?.focus();
      restoreFocus.current = false;
    }
  }, [draft]);

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!editable || draft === null || submitting.current) return;
    const name = draft.trim();
    if ([...name].length > 200 || name.includes('\u0000')) {
      setInvalid(true);
      return;
    }
    submitting.current = true;
    save.mutate(name || null);
  }

  return <section className="career-coordinator" aria-labelledby={`${fieldId}-title`}>
    <h3 id={`${fieldId}-title`}>Coordinador académico</h3>
    {coordinator.isPending && <p role="status">Cargando coordinador…</p>}
    {coordinator.isError && <div role="alert">
      <p>No se pudo consultar el nombre del coordinador.</p>
      <button type="button" className="shell-secondary" disabled={coordinator.isFetching}
        onClick={() => void coordinator.refetch()}>Reintentar consulta</button>
    </div>}
    {coordinator.data && <>
      {draft === null || !editable ? <>
        <p>{currentName ?? 'Sin coordinador registrado.'}</p>
        {editable && <button ref={editButton} type="button" className="shell-secondary" onClick={() => {
          setNotice('');
          setInvalid(false);
          save.reset();
          setDraft(currentName ?? '');
        }}>Editar coordinador</button>}
      </> : <form onSubmit={submit} aria-busy={save.isPending}>
        <label htmlFor={fieldId}>Nombre del coordinador</label>
        <input id={fieldId} autoFocus value={draft} readOnly={save.isPending}
          aria-invalid={invalid || undefined} aria-describedby={`${fieldId}-help`}
          onChange={event => { setDraft(event.target.value); setInvalid(false); save.reset(); }} />
        <p id={`${fieldId}-help`}>Hasta 200 caracteres. Deja el campo vacío para retirar el nombre.</p>
        {invalid && <p role="alert">Escribe un nombre válido de hasta 200 caracteres.</p>}
        {save.isError && <p role="alert">{saveError(save.error)}</p>}
        <div className="career-coordinator-actions">
          <button type="submit" className="shell-secondary" disabled={save.isPending}>
            {save.isPending ? 'Guardando…' : 'Guardar nombre'}
          </button>
          <button type="button" className="shell-secondary" disabled={save.isPending}
            onClick={() => { restoreFocus.current = true; setDraft(null); }}>Cancelar</button>
        </div>
      </form>}
      {notice && <p role="status">{notice}</p>}
    </>}
  </section>;
}
