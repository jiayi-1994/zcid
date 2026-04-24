import { useCallback, useEffect, useState } from 'react';
import type { PipelineSummary, PipelineConfig } from '../../services/pipeline';
import { fetchConnections, listBranches, type GitBranch, type GitConnection } from '../../services/integration';
import { ZModal } from '../ui/ZModal';
import { Btn } from '../ui/Btn';
import { Field } from '../ui/Field';
import { ZSelect } from '../ui/ZSelect';
import { ICode, IBranch } from '../ui/icons';

const DEFAULT_BRANCHES = ['main', 'master', 'develop', 'staging', 'release'];

interface RunPipelineModalProps {
  visible: boolean;
  pipeline: PipelineSummary | null;
  pipelineConfig?: PipelineConfig | null;
  onClose: () => void;
  onSubmit?: (params: { params?: Record<string, string>; gitBranch?: string; gitCommit?: string }) => Promise<void>;
}

const TRIGGER_LABELS: Record<string, string> = {
  manual: '手动触发',
  webhook: 'Webhook 触发',
  scheduled: '定时触发',
};

function extractRepoUrl(config?: PipelineConfig | null): string | undefined {
  if (!config?.stages) return undefined;
  for (const stage of config.stages) {
    for (const step of stage.steps) {
      if (step.type === 'git-clone' && step.config) {
        const url = step.config.repoUrl;
        if (typeof url === 'string' && url) return url;
      }
    }
  }
  return undefined;
}

function extractConfiguredBranch(config?: PipelineConfig | null): string | undefined {
  if (!config?.stages) return undefined;
  for (const stage of config.stages) {
    for (const step of stage.steps) {
      if (step.type === 'git-clone' && step.config) {
        const branch = step.config.branch;
        if (typeof branch === 'string' && branch) return branch;
      }
    }
  }
  return undefined;
}

function repoFullNameFromUrl(repoUrl: string): string | undefined {
  const m = repoUrl.match(/[:/]([^/:]+\/[^/.]+?)(?:\.git)?$/);
  return m?.[1];
}

function matchConnection(repoUrl: string, connections: GitConnection[]): GitConnection | undefined {
  try {
    const urlObj = new URL(repoUrl);
    return connections.find((c) => {
      try { return new URL(c.serverUrl).hostname === urlObj.hostname; }
      catch { return false; }
    });
  } catch {
    return undefined;
  }
}

export function RunPipelineModal({ visible, pipeline, pipelineConfig, onClose, onSubmit }: RunPipelineModalProps) {
  const [loading, setLoading] = useState(false);
  const [branchOptions, setBranchOptions] = useState<string[]>(DEFAULT_BRANCHES);
  const [branchLoading, setBranchLoading] = useState(false);
  const [gitBranch, setGitBranch] = useState('');
  const [gitCommit, setGitCommit] = useState('');

  const loadBranches = useCallback(async () => {
    const repoUrl = extractRepoUrl(pipelineConfig);
    const configured = extractConfiguredBranch(pipelineConfig);

    if (!repoUrl) {
      if (configured) {
        const opts = new Set(DEFAULT_BRANCHES);
        opts.add(configured);
        setBranchOptions(Array.from(opts));
      }
      return;
    }

    const repoFullName = repoFullNameFromUrl(repoUrl);
    if (!repoFullName) return;

    setBranchLoading(true);
    try {
      const connList = await fetchConnections();
      const conn = matchConnection(repoUrl, connList.items ?? []);
      if (!conn) {
        if (configured) {
          const opts = new Set(DEFAULT_BRANCHES);
          opts.add(configured);
          setBranchOptions(Array.from(opts));
        }
        return;
      }
      const branches = await listBranches(conn.id, repoFullName);
      if (branches.length > 0) {
        const names = branches.map((b: GitBranch) => b.name);
        const defaultBranch = branches.find((b: GitBranch) => b.isDefault);
        setBranchOptions(names);
        if (defaultBranch && !gitBranch) setGitBranch(defaultBranch.name);
      }
    } catch {
      if (configured) {
        const opts = new Set(DEFAULT_BRANCHES);
        opts.add(configured);
        setBranchOptions(Array.from(opts));
      }
    } finally {
      setBranchLoading(false);
    }
  }, [pipelineConfig, gitBranch]);

  useEffect(() => {
    if (visible) loadBranches();
  }, [visible, loadBranches]);

  useEffect(() => {
    if (!visible) { setGitBranch(''); setGitCommit(''); }
  }, [visible]);

  const handleSubmit = async () => {
    if (!onSubmit) return;
    setLoading(true);
    try {
      await onSubmit({
        gitBranch: gitBranch || undefined,
        gitCommit: gitCommit || undefined,
        params: {},
      });
    } catch {
      /* surfaced by caller */
    } finally {
      setLoading(false);
    }
  };

  if (!visible) return null;

  if (!onSubmit) {
    return (
      <ZModal
        title={`运行流水线: ${pipeline?.name ?? ''}`}
        onClose={onClose}
        footer={<Btn onClick={onClose}>关闭</Btn>}
      >
        <div style={{ color: 'var(--z-500)', fontSize: 13 }}>功能开发中（Pipeline 执行属于 Epic 7）</div>
      </ZModal>
    );
  }

  return (
    <ZModal
      title=""
      onClose={onClose}
      footer={
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
          <Btn onClick={onClose}>取消</Btn>
          <Btn variant="primary" icon={<ICode size={12} />} onClick={handleSubmit} disabled={loading}>
            {loading ? '触发中...' : '开始运行'}
          </Btn>
        </div>
      }
    >
      <div style={{ marginBottom: 18, display: 'flex', alignItems: 'center', gap: 12 }}>
        <div style={{
          width: 40, height: 40, borderRadius: 10,
          background: 'linear-gradient(135deg, var(--accent-1), var(--accent-2))',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          color: '#fff', flex: 'none',
        }}>
          <ICode size={18} />
        </div>
        <div>
          <div style={{ fontSize: 14, fontWeight: 600, color: 'var(--z-900)' }}>
            运行 {pipeline?.name ?? '流水线'}
          </div>
          <div className="sub" style={{ fontSize: 12 }}>
            {TRIGGER_LABELS[pipeline?.triggerType ?? ''] || '手动触发'} · {pipeline?.concurrencyPolicy === 'reject' ? '拒绝并发' : '队列模式'}
          </div>
        </div>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        <Field label={<span style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}><IBranch size={12} /> 构建分支</span>}>
          {branchLoading ? (
            <div style={{ fontSize: 12, color: 'var(--z-500)' }}>加载分支中...</div>
          ) : (
            <ZSelect
              width={320}
              value={gitBranch}
              options={[{ value: '', label: '选择或输入分支名' }, ...branchOptions.map((b) => ({ value: b, label: b }))]}
              onChange={setGitBranch}
            />
          )}
        </Field>
        <Field label="Git Commit（可选）" help="指定 commit SHA，留空则使用分支最新">
          <input
            className="input mono"
            value={gitCommit}
            onChange={(e) => setGitCommit(e.target.value)}
            placeholder="abc1234..."
          />
        </Field>
      </div>
    </ZModal>
  );
}
