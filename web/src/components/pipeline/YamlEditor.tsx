import { useEffect, useRef, useState, useCallback } from 'react';
import { Input } from '@arco-design/web-react';
import type { PipelineConfig } from '../../services/pipeline';
import { configToJson, jsonToConfig } from '../../lib/pipeline/configJson';

const { TextArea } = Input;

interface YamlEditorProps {
  config: PipelineConfig;
  onChange: (config: PipelineConfig) => void;
  onValidationError?: (msg: string) => void;
  disabled?: boolean;
}

const DEBOUNCE_MS = 300;

export function YamlEditor({ config, onChange, onValidationError, disabled }: YamlEditorProps) {
  const [value, setValue] = useState(() => configToJson(config));
  const [error, setError] = useState<string | null>(null);
  const internalChange = useRef(false);
  const debounceTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;
  const onValidationErrorRef = useRef(onValidationError);
  onValidationErrorRef.current = onValidationError;

  useEffect(() => {
    if (internalChange.current) {
      internalChange.current = false;
      return;
    }
    setValue(configToJson(config));
  }, [config]);

  useEffect(() => {
    return () => { if (debounceTimer.current) clearTimeout(debounceTimer.current); };
  }, []);

  const handleChange = useCallback((v: string) => {
    setValue(v);
    setError(null);
    if (debounceTimer.current) clearTimeout(debounceTimer.current);
    debounceTimer.current = setTimeout(() => {
      try {
        const parsed = jsonToConfig(v);
        internalChange.current = true;
        onChangeRef.current(parsed);
      } catch (e) {
        const msg = e instanceof Error ? e.message : '解析失败';
        setError(msg);
        onValidationErrorRef.current?.(msg);
      }
    }, DEBOUNCE_MS);
  }, []);

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
