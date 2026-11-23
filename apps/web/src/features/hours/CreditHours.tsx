import { useId, useRef, type FormEvent } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api, ApiError } from '../../lib/api';

const calendar = new Intl.DateTimeFormat('en-CA', {
  dateStyle: 'short', timeZone: 'America/El_Salvador',
});

function message(error: Error) {
  if (error instanceof ApiError) {
    switch (error.field) {
      case 'fecha_horas': return 'Usa una fecha desde la asignación hasta hoy.';
      case 'estado_participacion': return 'La participación o el proyecto ya está cerrado.';
      case 'horas': return 'Ingresa una cantidad entera y positiva.';
      case 'concepto': return 'Escribe un concepto de hasta 500 caracteres.';
      case 'solicitud_id': return 'La solicitud ya existe con otros datos. Revisa el libro antes de iniciar un nuevo registro.';
      case 'escritura_core': return 'La acreditación aún no está habilitada en este entorno.';
    }
  }
  return 'No se pudo confirmar el registro. Usa Reintentar para enviar los mismos datos con la misma clave.';
}

type Input = { horas: number; fecha_horas: string; concepto: string; solicitud_id: string };

export default function CreditHours({ project, enrollment, student, assigned }: {
  project: string; enrollment: string; student: string; assigned: string;
}) {
  const title = useId();
  const dialog = useRef<HTMLDialogElement>(null);
  const last = useRef<Input | undefined>(undefined);
  const form = useRef<HTMLFormElement>(null);
  const client = useQueryClient();
  const save = useMutation({
    mutationFn: (body: Input) => api.request<{ id: string; repetida: boolean }>(
      `/api/v1/proyectos/${encodeURIComponent(project)}/participaciones/${encodeURIComponent(enrollment)}/horas`,
      { method: 'POST', body }),
    onSuccess: async () => {
      form.current?.reset();
      last.current = undefined;
      dialog.current?.close();
      await client.invalidateQueries({ queryKey: ['hours', student] });
    },
  });

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (save.isPending) return;
    const data = new FormData(event.currentTarget);
    const input = {
      horas: Number(data.get('horas')), fecha_horas: String(data.get('fecha_horas')),
      concepto: String(data.get('concepto') ?? '').trim(),
    };
    const old = last.current;
    const same = old && old.horas === input.horas && old.fecha_horas === input.fecha_horas && old.concepto === input.concepto;
    const body = { ...input, solicitud_id: same ? old.solicitud_id : crypto.randomUUID() };
    last.current = body;
    save.mutate(body);
  }

  return <>
    <button type="button" className="shell-secondary" onClick={() => dialog.current?.showModal()}>Acreditar horas</button>
    {save.isSuccess && <span role="status"> Horas acreditadas.</span>}
    <dialog ref={dialog} className="shell-confirm" aria-labelledby={title}
      onCancel={event => { if (save.isPending) event.preventDefault(); }}>
      <h2 id={title}>Acreditar horas</h2>
      <p>El movimiento conserva la naturaleza vigente del proyecto al acreditarse.</p>
      <form ref={form} className="institution-form" onSubmit={submit}>
        <label>Horas <input name="horas" type="number" min={1} max={2147483647} step={1} required disabled={save.isPending} /></label>
        <label>Fecha de actividad <input name="fecha_horas" type="date" min={calendar.format(new Date(assigned))}
          max={calendar.format(new Date())} defaultValue={calendar.format(new Date())} required disabled={save.isPending} /></label>
        <label>Concepto <textarea name="concepto" maxLength={500} required disabled={save.isPending} /></label>
        {save.error && <p role="alert" className="access-error">{message(save.error)}</p>}
        <div className="institution-actions">
          <button type="button" className="shell-secondary" disabled={save.isPending} onClick={() => dialog.current?.close()}>Volver</button>
          <button type="submit" className="access-primary" disabled={save.isPending}>{save.isPending ? 'Guardando…' : 'Acreditar'}</button>
          {save.error && last.current && <button type="button" className="shell-secondary" onClick={() => {
            if (last.current) save.mutate(last.current);
          }}>Reintentar</button>}
        </div>
      </form>
    </dialog>
  </>;
}
