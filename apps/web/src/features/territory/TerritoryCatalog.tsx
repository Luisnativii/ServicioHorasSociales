import { useState } from 'react';
import TerritorySelector, { type TerritorySelection } from './TerritorySelector';

export default function TerritoryCatalog() {
  const [value, setValue] = useState<TerritorySelection>({
    departmentId: '', municipalityId: '', districtId: '',
  });

  return <section className="accounts-page">
    <h1>Catálogo territorial</h1>
    <TerritorySelector value={value} onChange={setValue} />
  </section>;
}
