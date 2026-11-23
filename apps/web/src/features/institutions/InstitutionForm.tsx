import { useState, type FormEvent } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ApiError, NetworkError } from '../../lib/api';
import TerritorySelector, { type TerritorySelection } from '../territory/TerritorySelector';
import { institutionKey, saveInstitution, type Institution } from './api';

const emptyTerritory = { departmentId: '', municipalityId: '', districtId: '' };
const contactFields = [
  { key: 'nombre', label: 'Nombre del contacto', type: 'text', limit: 200 },
  { key: 'correo', label: 'Correo', type: 'email', limit: undefined },
  { key: 'telefono', label: 'Teléfono', type: 'tel', limit: 20 },
  { key: 'extension', label: 'Extensión', type: 'text', limit: 200 },
] as const;

export function institutionError(error: Error) {
  if (error instanceof NetworkError) return 'Se perdió la conexión. Consulta el registro antes de intentar guardar otra vez.';
  if (error instanceof ApiError && error.status === 409) return 'Ya existe una institución con ese nombre.';
  if (error instanceof ApiError && error.status === 400) return 'Revisa el nombre, la sede y los datos del contacto.';
  if (error instanceof ApiError && error.status === 404) return 'La institución ya no está disponible.';
  return 'No se pudo completar la operación. Inténtalo de nuevo.';
}

type Props = {
  institution?: Institution;
  onCancel: () => void;
  onSaved: (institution: Institution) => void;
};

export default function InstitutionForm({ institution, onCancel, onSaved }: Props) {
  const client = useQueryClient();
  const [replaceSeat, setReplaceSeat] = useState(false);
  const [editContact, setEditContact] = useState(false);
  const [territory, setTerritory] = useState<TerritorySelection>(emptyTerritory);
  const [validation, setValidation] = useState('');
  const save = useMutation({
    mutationFn: (input: Parameters<typeof saveInstitution>[0]) => saveInstitution(input, institution?.id),
    onSuccess: result => {
      void client.invalidateQueries({ queryKey: institutionKey });
      client.setQueryData([...institutionKey, 'detail', result.id], result);
      onSaved(result);
    },
  });

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (save.isPending) return;
    save.reset();
    const form = new FormData(event.currentTarget);
    const value = (key: string) => String(form.get(key) ?? '').trim();
    const nombre = value('nombre');
    const distrito_id = replaceSeat || !institution?.sede
      ? territory.districtId || null
      : institution.sede.distrito_id;
    if (!nombre) {
      setValidation('Escribe el nombre de la institución.');
      return;
    }
    if ((replaceSeat || !institution?.sede) && territory.departmentId && !territory.districtId) {
      setValidation('Selecciona un distrito o deja el departamento sin seleccionar.');
      return;
    }
    const contacto = editContact ? {
      nombre: value('contacto-nombre'),
      correo: value('contacto-correo') || null,
      telefono: value('contacto-telefono') || null,
      extension: value('contacto-extension') || null,
    } : undefined;
    if (contacto && (!contacto.nombre ||
      (contacto.telefono && !/^(?=.*[0-9])[0-9+\-\s]{8,20}$/.test(contacto.telefono)) ||
      (contacto.extension && !contacto.telefono))) {
      setValidation('Escribe el nombre del contacto y, si corresponde, un teléfono de 8 a 20 caracteres. La extensión requiere teléfono.');
      return;
    }
    setValidation('');
    save.mutate({ nombre, distrito_id, contacto });
  }

  return <section className="institution-panel" aria-labelledby="institution-form-title">
    <h2 id="institution-form-title" tabIndex={-1}>{institution ? 'Editar institución' : 'Nueva institución'}</h2>
    <form className="institution-form" onSubmit={submit}>
      <fieldset disabled={save.isPending}>
        <label className="institution-field">Nombre de la institución
          <input name="nombre" defaultValue={institution?.nombre} required maxLength={200} />
        </label>
        {institution?.sede && <>
          <p>Distrito guardado: {institution.sede.nombre}</p>
          {!replaceSeat && institution.ubicacion_heredada &&
            <p>Ubicación histórica: {institution.ubicacion_heredada}</p>}
          <label className="institution-toggle">
            <input type="checkbox" checked={replaceSeat} onChange={event => {
              setReplaceSeat(event.target.checked);
              setTerritory(emptyTerritory);
            }} />
            Cambiar o quitar la sede
          </label>
          {replaceSeat && <p>Deja el departamento sin seleccionar para quitar la sede.</p>}
        </>}
        {(!institution?.sede || replaceSeat) && <TerritorySelector value={territory}
          onChange={setTerritory} historicalName={institution?.ubicacion_heredada} />}
        <fieldset className="institution-contact">
          <legend>Contacto principal</legend>
          <label className="institution-toggle">
            <input type="checkbox" checked={editContact} onChange={event => setEditContact(event.target.checked)} />
            {institution?.contacto ? 'Editar contacto principal' : 'Registrar contacto principal'}
          </label>
          <div hidden={!editContact}>
            {contactFields.map(field => <label className="institution-field" key={field.key}>
              {field.label}
              <input name={`contacto-${field.key}`} type={field.type}
                defaultValue={institution?.contacto?.[field.key] ?? ''}
                maxLength={field.limit} disabled={!editContact} required={editContact && field.key === 'nombre'} />
            </label>)}
          </div>
        </fieldset>
      </fieldset>
      {validation && <p className="access-error" role="alert">{validation}</p>}
      {save.error && <p className="access-error" role="alert">{institutionError(save.error)}</p>}
      <div className="institution-actions">
        <button type="submit" className="access-primary" disabled={save.isPending}>{save.isPending ? 'Guardando…' : 'Guardar institución'}</button>
        <button type="button" className="shell-secondary" disabled={save.isPending} onClick={onCancel}>Cancelar</button>
      </div>
    </form>
  </section>;
}
