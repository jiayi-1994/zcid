import type { PipelineConfig } from '../../services/pipeline';

/**
 * Convert PipelineConfig to JSON string (pretty-printed).
 * Used for "JSON 模式" editing in the pipeline editor.
 */
export function configToJson(config: PipelineConfig): string {
  return JSON.stringify(config, null, 2);
}

/**
 * Parse JSON string back to PipelineConfig.
 * @throws Error if JSON is invalid or structure is invalid
 */
export function jsonToConfig(jsonText: string): PipelineConfig {
  let parsed: unknown;
  try {
    parsed = JSON.parse(jsonText);
  } catch (e) {
    throw new Error('无效的 JSON 格式');
  }

  if (!parsed || typeof parsed !== 'object') {
    throw new Error('配置必须是对象');
  }

  const obj = parsed as Record<string, unknown>;
  if (!obj.schemaVersion || typeof obj.schemaVersion !== 'string') {
    throw new Error('缺少 schemaVersion 字段');
  }
  if (!Array.isArray(obj.stages)) {
    throw new Error('缺少或无效的 stages 字段（应为数组）');
  }

  for (let i = 0; i < (obj.stages as unknown[]).length; i++) {
    const s = (obj.stages as Record<string, unknown>[])[i];
    if (!s || typeof s !== 'object') {
      throw new Error(`stages[${i}] 必须是对象`);
    }
    if (typeof s.id !== 'string' || !s.id) {
      throw new Error(`stages[${i}] 缺少有效的 id 字段`);
    }
    if (typeof s.name !== 'string' || !s.name) {
      throw new Error(`stages[${i}] 缺少有效的 name 字段`);
    }
    if (!Array.isArray(s.steps)) {
      throw new Error(`stages[${i}] 缺少有效的 steps 数组`);
    }
  }

  return parsed as PipelineConfig;
}
