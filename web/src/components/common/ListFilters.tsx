import { ZSelect } from '../ui/ZSelect';
import { ISearch } from '../ui/icons';

export interface FilterOption {
  label: string;
  value: string;
}

export interface FilterConfig {
  key: string;
  type: 'search' | 'select';
  placeholder?: string;
  options?: FilterOption[];
}

interface ListFiltersProps<T extends Record<string, string>> {
  filters: FilterConfig[];
  values: T;
  onChange: (updates: Partial<T>) => void;
}

export function ListFilters<T extends Record<string, string>>({ filters, values, onChange }: ListFiltersProps<T>) {
  return (
    <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
      {filters.map((f) => {
        if (f.type === 'search') {
          return (
            <div key={f.key} className="input-wrap">
              <ISearch size={13} />
              <input
                className="input input--with-icon"
                style={{ width: 200 }}
                placeholder={f.placeholder ?? '搜索'}
                value={values[f.key] ?? ''}
                onChange={(e) => onChange({ [f.key]: e.target.value } as Partial<T>)}
              />
            </div>
          );
        }
        return (
          <ZSelect
            key={f.key}
            width={140}
            value={values[f.key] ?? ''}
            options={[{ value: '', label: f.placeholder ?? '全部' }, ...(f.options ?? [])]}
            onChange={(v) => onChange({ [f.key]: v ?? '' } as Partial<T>)}
          />
        );
      })}
    </div>
  );
}
