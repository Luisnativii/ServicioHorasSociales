import { useId, useRef, type FormEvent } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api, ApiError } from '../../lib/api';

type Action = 'cancelar' | 'retirar' | 'finalizar';
const today = () => new Intl.DateTimeFormat('en-CA', {
  dateStyle: 'short', timeZone: 'America/El_Salvador',
}).format(new Date());

function failure(error: Error) {
  if (error instanceof ApiError) {
    switch (error.field) {
      case 'fecha_cierre': return 'La fecha no puede ser futura ni anterior al inicio, asignación u horas registradas.';
      case 'participaciones_abiertas': return 'Para finalizar, el proyecto debe tener participaciones y todas deben estar cerradas.';
      case 'estado_proyecto': return 'El proyecto ya no está activo. Vuelve a consultar su estado.';
      case 'estado_participacion': return 'La participación ya está cerrada.';
      case 'escritura_core': return 'Los cierres aún no están habilitados en este entorno.';
      case 'motivo': return 'Escribe un motivo de hasta 500 caracteres.';
    }
  }
  return 'No se pudo confirmar el cierre. Consulta el estado antes de volver a intentarlo.';
}

export default function CloseAction({ project, enrollment, action }: { project: string; enrollment?: string; action: Action }) {
  const title = useId();
  const dialog = useRef<HTMLDialogElement>(null);
  const form = useRef<HTMLFormElement>(null);
  const client = useQueryClient();
  const label = action === 'retirar' ? 'Retirar estudiante'
    : action === 'cancelar' ? 'Cancelar proyecto'
      : enrollment ? 'Finalizar participación' : 'Finalizar proyecto';
  const required = action !== 'finalizar';
  const close = useMutation({
    mutationFn: (body: { accion: Action; fecha_cierre: string; motivo: string }) => api.request(
      `/api/v1/proyectos/${encodeURIComponent(project)}${enrollment ? `/participaciones/${encodeURIComponent(enrollment)}` : ''}/cierre`,
      { method: 'POST', body }),
    onSuccess: async () => {
      dialog.current?.close();
      await Promise.all([
        client.invalidateQueries({ queryKey: ['participations'] }),
        client.invalidateQueries({ queryKey: ['projects'] }),
        client.invalidateQueries({ queryKey: ['participation-candidates'] }),
      ]);
    },
    onError: error => {
      if (error instanceof ApiError && ['estado_proyecto', 'estado_participacion'].includes(error.field ?? '')) {
        // Keep the conflict visible until the user closes the dialog.
        void client.invalidateQueries({ queryKey: ['projects'], refetchType: 'none' });
        void client.invalidateQueries({ queryKey: ['participations'], refetchType: 'none' });
      }
    },
  });

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (close.isPending) return;
    const data = new FormData(event.currentTarget);
    close.mutate({ accion: action, fecha_cierre: String(data.get('fecha_cierre')), motivo: String(data.get('motivo') ?? '') });
  }

  return <>
    <button type="button" className="shell-secondary" onClick={() => {
      close.reset();
      form.current?.reset();
      dialog.current?.showModal();
    }}>{label}</button>
    <dialog ref={dialog} className="shell-confirm" aria-labelledby={title}
      onCancel={event => { if (close.isPending) event.preventDefault(); }}
      onClose={() => {
        if (close.error instanceof ApiError && ['estado_proyecto', 'estado_participacion'].includes(close.error.field ?? '')) {
          void client.invalidateQueries({ queryKey: ['projects'] });
          void client.invalidateQueries({ queryKey: ['participations'] });
        }
      }}>
      <h2 id={title}>{label}</h2>
      <p>{action === 'cancelar' ? 'Se cancelarán las participaciones abiertas. Las horas y las participaciones ya cerradas se conservan.'
        : action === 'retirar' ? 'Se conservarán las horas y se bloqueará la reinscripción. Si todas las participaciones quedan retiradas, el proyecto se marcará abandonado.'
          : enrollment ? 'La participación quedará cerrada y conservará las horas registradas.'
            : 'El proyecto debe tener al menos una participación y todas deben estar cerradas.'}</p>
      <form ref={form} className="institution-form" onSubmit={submit}>
        <label>Fecha de cierre <input name="fecha_cierre" type="date" max={today()} defaultValue={today()} required disabled={close.isPending} /></label>
        <label>{required ? 'Motivo' : 'Motivo (opcional)'} <textarea name="motivo" maxLength={500} required={required} disabled={close.isPending} /></label>
        {close.error && <p role="alert" className="access-error">{failure(close.error)}</p>}
        <div className="institution-actions">
          <button type="button" className="shell-secondary" disabled={close.isPending} onClick={() => dialog.current?.close()}>Volver sin guardar</button>
          <button type="submit" className="access-primary" disabled={close.isPending}>{close.isPending ? 'Guardando…' : 'Confirmar cierre'}</button>
        </div>
      </form>
    </dialog>
  </>;
}
