import { useId } from 'react';
import { useQuery } from '@tanstack/react-query';
import { listTerritory, type TerritoryLevel } from './api';

export type TerritorySelection = {
  departmentId: string;
  municipalityId: string;
  districtId: string;
};

type TerritorySelectorProps = {
  value: TerritorySelection;
  onChange: (value: TerritorySelection) => void;
  historicalName?: string | null;
};

function useTerritoryOptions(level: TerritoryLevel, parentId = '', enabled = true) {
  return useQuery({
    queryKey: ['territory-catalog', level, parentId],
    queryFn: ({ signal }) => listTerritory(level, parentId, signal),
    enabled,
  });
}

type TerritoryFieldProps = {
  label: string;
  value: string;
  onChange: (value: string) => void;
  query: ReturnType<typeof useTerritoryOptions>;
  prerequisite?: string;
};

function TerritoryField({ label, value, onChange, query, prerequisite }: TerritoryFieldProps) {
  const id = useId();
  const options = query.data ?? [];
  const unavailable = Boolean(value) && !options.some(option => option.id === value);
  const loading = !prerequisite && query.isPending && !query.isError;
  const failure = !prerequisite && query.isError;
  const message = prerequisite ?? (
    loading ? 'Cargando opciones…' :
    failure ? 'No se pudieron cargar las opciones.' :
    unavailable ? 'La selección guardada no está disponible en este catálogo.' :
    options.length === 0 ? 'No hay opciones disponibles.' : null
  );

  return <div className="territory-field">
    <label htmlFor={id}>{label}</label>
    <select
      id={id}
      value={value}
      disabled={Boolean(prerequisite) || loading || (!query.data && failure) || (!value && options.length === 0)}
      aria-describedby={message ? `${id}-info` : undefined}
      onChange={event => onChange(event.target.value)}
    >
      <option value="">Sin seleccionar</option>
      {unavailable && <option value={value} disabled>Selección guardada</option>}
      {options.map(option => <option key={option.id} value={option.id}>{option.nombre}</option>)}
    </select>
    {message && <p id={`${id}-info`} role={failure ? 'alert' : 'status'}>{message}</p>}
    {failure && <button type="button" className="shell-secondary" disabled={query.isFetching}
      onClick={() => void query.refetch()}>Reintentar carga de {label.toLowerCase()}</button>}
  </div>;
}

export default function TerritorySelector({ value, onChange, historicalName }: TerritorySelectorProps) {
  const departments = useTerritoryOptions('departamentos');
  const departmentAvailable = departments.data?.some(place => place.id === value.departmentId) ?? false;
  const municipalities = useTerritoryOptions('municipios', value.departmentId, departmentAvailable);
  const municipalityAvailable = departmentAvailable &&
    (municipalities.data?.some(place => place.id === value.municipalityId) ?? false);
  const districts = useTerritoryOptions('distritos', value.municipalityId, municipalityAvailable);

  return <fieldset className="territory-selector">
    <legend>Ubicación territorial</legend>
    <div className="territory-fields">
      <TerritoryField label="Departamento" value={value.departmentId} query={departments}
        onChange={departmentId => onChange({ departmentId, municipalityId: '', districtId: '' })} />
      <TerritoryField label="Municipio" value={value.municipalityId} query={municipalities}
        prerequisite={departmentAvailable ? undefined : 'Selecciona un departamento disponible.'}
        onChange={municipalityId => onChange({ ...value, municipalityId, districtId: '' })} />
      <TerritoryField label="Distrito" value={value.districtId} query={districts}
        prerequisite={municipalityAvailable ? undefined : 'Selecciona un municipio disponible.'}
        onChange={districtId => onChange({ ...value, districtId })} />
    </div>
    {!value.districtId && <p className="territory-note">Distrito sin identificar.</p>}
    {historicalName?.trim() && <p className="territory-note">Nombre histórico registrado: {historicalName}</p>}
  </fieldset>;
}
