import { Input, Select, Space } from '@arco-design/web-react';

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

export function ListFilters<T extends Record<string, string>>({
  filters,
  values,
  onChange,
}: ListFiltersProps<T>) {
  return (
    <Space wrap>
      {filters.map((f) => {
        if (f.type === 'search') {
          return (
            <Input.Search
              key={f.key}
              placeholder={f.placeholder ?? '搜索'}
              style={{ width: 200 }}
              value={values[f.key] ?? ''}
              onChange={(v) => onChange({ [f.key]: v } as Partial<T>)}
              onSearch={(v) => onChange({ [f.key]: v } as Partial<T>)}
              allowClear
            />
          );
        }
        return (
          <Select
            key={f.key}
            placeholder={f.placeholder}
            style={{ width: 140 }}
            value={values[f.key] ?? ''}
            onChange={(v) => onChange({ [f.key]: v ?? '' } as Partial<T>)}
            options={f.options ?? []}
            allowClear
          />
        );
      })}
    </Space>
  );
}
