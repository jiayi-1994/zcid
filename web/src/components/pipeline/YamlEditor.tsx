import { useCallback, useEffect, useRef, useState } from 'react';
import type { PipelineConfig } from '../../services/pipeline';
import { configToJson, jsonToConfig } from '../../lib/pipeline/configJson';

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
    if (internalChange.current) { internalChange.current = false; return; }
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
    <div className="zc" style={{ height: '100%', display: 'flex', flexDirection: 'column', fontFamily: 'var(--font-sans)' }}>
      <textarea
        className="input mono"
        value={value}
        onChange={(e) => handleChange(e.target.value)}
        disabled={disabled}
        placeholder='{"schemaVersion":"1.0","stages":[]}'
        spellCheck={false}
        style={{
          flex: 1,
          height: '100%',
          minHeight: 300,
          fontSize: 12.5,
          lineHeight: 1.6,
          resize: 'none',
          padding: 12,
        }}
      />
      {error && (
        <div role="alert" style={{ marginTop: 8, color: 'var(--red-ink)', background: 'var(--red-soft)', padding: '6px 10px', borderRadius: 6, fontSize: 12, fontFamily: 'var(--font-mono)' }}>
          {error}
        </div>
      )}
    </div>
  );
}
