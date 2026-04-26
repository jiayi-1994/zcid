import { useEffect, useState } from 'react';
import { Message } from '@arco-design/web-react';
import { Btn } from '../../components/ui/Btn';
import { Field } from '../../components/ui/Field';
import { ZModal } from '../../components/ui/ZModal';
import { ZSelect } from '../../components/ui/ZSelect';
import { createAccessToken, type AccessTokenRecord } from '../../services/accessToken';
import { extractErrorMessage } from '../../services/http';

const SCOPE_GROUPS = [
  {
    name: 'Pipelines',
    scopes: [
      { value: 'pipelines:read', label: '读取流水线与运行', description: '查看流水线、运行记录和步骤时间线' },
      { value: 'pipelines:trigger', label: '触发流水线', description: '创建新的 Pipeline Run' },
    ],
  },
  {
    name: 'Deployments',
    scopes: [
      { value: 'deployments:read', label: '读取部署', description: '查看部署和同步状态' },
      { value: 'deployments:write', label: '写入部署', description: '创建、更新或删除部署' },
    ],
  },
  {
    name: 'Variables',
    scopes: [{ value: 'variables:read', label: '读取变量', description: '读取非敏感变量元数据' }],
  },
  {
    name: 'Notifications',
    scopes: [{ value: 'notifications:read', label: '读取通知规则', description: '查看项目通知规则' }],
  },
  {
    name: 'Admin',
    scopes: [{ value: 'admin:read', label: '读取管理信息', description: '查看只读管理 API（管理员令牌）' }],
  },
];

interface TokenFormModalProps {
  visible: boolean;
  onClose: () => void;
  onCreated: (record: AccessTokenRecord, rawToken: string) => void;
}

export function TokenFormModal({ visible, onClose, onCreated }: TokenFormModalProps) {
  const [name, setName] = useState('');
  const [type, setType] = useState<'personal' | 'project'>('personal');
  const [projectId, setProjectId] = useState('');
  const [expiresAt, setExpiresAt] = useState('');
  const [scopes, setScopes] = useState<string[]>(['pipelines:trigger']);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!visible) return;
    const defaultExpiry = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000);
    setName('');
    setType('personal');
    setProjectId('');
    setExpiresAt(defaultExpiry.toISOString().slice(0, 16));
    setScopes(['pipelines:read', 'pipelines:trigger']);
  }, [visible]);

  if (!visible) return null;

  const toggleScope = (scope: string) => {
    setScopes((prev) => prev.includes(scope) ? prev.filter((s) => s !== scope) : [...prev, scope]);
  };

  const toggleGroup = (groupScopes: string[]) => {
    setScopes((prev) => {
      const allSelected = groupScopes.every((scope) => prev.includes(scope));
      if (allSelected) return prev.filter((scope) => !groupScopes.includes(scope));
      return Array.from(new Set([...prev, ...groupScopes]));
    });
  };

  const submit = async () => {
    if (!name.trim()) { Message.error('请输入令牌名称'); return; }
    if (scopes.length === 0) { Message.error('至少选择一个 scope'); return; }
    if (type === 'project' && !projectId.trim()) { Message.error('项目令牌需要项目 ID'); return; }
    setLoading(true);
    try {
      const result = await createAccessToken({
        name: name.trim(),
        type,
        scopes,
        expiresAt: new Date(expiresAt).toISOString(),
        projectId: type === 'project' ? projectId.trim() : undefined,
      });
      onCreated(result.token, result.rawToken);
      onClose();
    } catch (error: unknown) {
      Message.error(extractErrorMessage(error, '创建令牌失败'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <ZModal
      title="新建访问令牌"
      onClose={onClose}
      footer={
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
          <Btn onClick={onClose}>取消</Btn>
          <Btn variant="primary" onClick={submit} disabled={loading}>{loading ? '创建中...' : '创建令牌'}</Btn>
        </div>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        <div className="notice" style={{ padding: 12, borderRadius: 10, background: 'var(--z-50)', color: 'var(--z-600)', fontSize: 12 }}>
          令牌只会在创建后显示一次。请复制到密码管理器或 CI Secret 中，后续列表不会再次展示明文。
        </div>
        <Field label="名称" required>
          <input className="input" value={name} onChange={(e) => setName(e.target.value)} placeholder="例如 GitHub Actions 发布流水线" />
        </Field>
        <Field label="类型" required>
          <ZSelect
            width={240}
            value={type}
            options={[{ value: 'personal', label: '个人访问令牌' }, { value: 'project', label: '项目访问令牌' }]}
            onChange={(v) => setType(v as 'personal' | 'project')}
          />
        </Field>
        {type === 'project' && (
          <Field label="项目 ID" required help="项目令牌只能访问这个项目。">
            <input className="input" value={projectId} onChange={(e) => setProjectId(e.target.value)} placeholder="project uuid" />
          </Field>
        )}
        <Field label="过期时间" required>
          <input className="input" type="datetime-local" value={expiresAt} onChange={(e) => setExpiresAt(e.target.value)} />
        </Field>
        <Field label="Scopes" required>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            {SCOPE_GROUPS.map((group) => {
              const groupValues = group.scopes.map((scope) => scope.value);
              const allSelected = groupValues.every((scope) => scopes.includes(scope));
              return (
                <div key={group.name} style={{ padding: 10, border: '1px solid var(--z-150)', borderRadius: 8, background: 'var(--z-25)' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', gap: 10, alignItems: 'center', marginBottom: 8 }}>
                    <span style={{ fontSize: 12, fontWeight: 650 }}>{group.name}</span>
                    <button type="button" className={`btn btn--xs ${allSelected ? 'btn--primary' : 'btn--outline'}`} onClick={() => toggleGroup(groupValues)}>
                      {allSelected ? '取消全选' : '全选'}
                    </button>
                  </div>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                    {group.scopes.map((scope) => {
                      const selected = scopes.includes(scope.value);
                      return (
                        <button
                          key={scope.value}
                          type="button"
                          className={`btn ${selected ? 'btn--primary' : 'btn--outline'}`}
                          onClick={() => toggleScope(scope.value)}
                          style={{ height: 'auto', minHeight: 30, justifyContent: 'space-between', textAlign: 'left', whiteSpace: 'normal' }}
                        >
                          <span style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                            <b style={{ fontWeight: 600 }}>{scope.label}</b>
                            <small style={{ opacity: selected ? 0.86 : 0.7 }}>{scope.description}</small>
                          </span>
                          <span className="mono" style={{ fontSize: 10.5 }}>{scope.value}</span>
                        </button>
                      );
                    })}
                  </div>
                </div>
              );
            })}
          </div>
        </Field>
      </div>
    </ZModal>
  );
}
