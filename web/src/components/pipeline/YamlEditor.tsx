import { useEffect, useRef, useState } from 'react';
import { Input } from '@arco-design/web-react';
import type { PipelineConfig } from '../../services/pipeline';
import { configToJson, jsonToConfig } from '../../lib/pipeline/configYaml';

const { TextArea } = Input;

interface YamlEditorProps {
  config: PipelineConfig;
  onChange: (config: PipelineConfig) => void;
  onValidationError?: (msg: string) => void;
  disabled?: boolean;
}

export function YamlEditor({ config, onChange, onValidationError, disabled }: YamlEditorProps) {
  const [value, setValue] = useState(() => configToJson(config));
  const [error, setError] = useState<string | null>(null);
  const internalChange = useRef(false);

  useEffect(() => {
    if (internalChange.current) {
      internalChange.current = false;
      return;
    }
    setValue(configToJson(config));
  }, [config]);

  const handleChange = (v: string) => {
    setValue(v);
    setError(null);
    try {
      const parsed = jsonToConfig(v);
      internalChange.current = true;
      onChange(parsed);
    } catch (e) {
      const msg = e instanceof Error ? e.message : '解析失败';
      setError(msg);
      onValidationError?.(msg);
    }
  };

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <TextArea
        value={value}
        onChange={handleChange}
        disabled={disabled}
        placeholder='{"schemaVersion":"1.0","stages":[]}'
        style={{
          flex: 1,
          fontFamily: 'ui-monospace, "Cascadia Code", "Source Code Pro", Menlo, Monaco, monospace',
          fontSize: 13,
          lineHeight: 1.5,
          resize: 'none',
        }}
        autoSize={{ minRows: 10 }}
      />
      {error && (
        <div role="alert" style={{ marginTop: 8, color: 'var(--color-red-6)', fontSize: 12 }}>
          {error}
        </div>
      )}
    </div>
  );
}
