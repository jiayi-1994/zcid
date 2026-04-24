import { useEffect, useState } from 'react';
import { Message } from '@arco-design/web-react';
import { ZModal } from '../../../components/ui/ZModal';
import { Btn } from '../../../components/ui/Btn';
import { Field } from '../../../components/ui/Field';
import { ZSelect } from '../../../components/ui/ZSelect';

interface ConnectionFormModalProps {
  visible: boolean;
  onClose: () => void;
  onSubmit: (data: {
    name: string;
    providerType: string;
    serverUrl: string;
    accessToken: string;
    description: string;
  }) => Promise<void>;
  editMode?: boolean;
  initialValues?: { name?: string; description?: string };
}

const EMPTY = { name: '', providerType: 'gitlab', serverUrl: '', accessToken: '', description: '' };

export function ConnectionFormModal({ visible, onClose, onSubmit, editMode = false, initialValues }: ConnectionFormModalProps) {
  const [form, setForm] = useState(EMPTY);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (visible) {
      setForm({ ...EMPTY, name: initialValues?.name ?? '', description: initialValues?.description ?? '' });
    }
  }, [visible, initialValues]);

  const handleOk = async () => {
    if (!form.name.trim()) { Message.error('请输入连接名称'); return; }
    if (!editMode) {
      if (!form.serverUrl.trim()) { Message.error('请输入 Server URL'); return; }
      if (!form.accessToken) { Message.error('请输入 Access Token'); return; }
    }
    setSubmitting(true);
    try {
      await onSubmit({
        name: form.name.trim(),
        providerType: form.providerType,
        serverUrl: form.serverUrl.trim(),
        accessToken: form.accessToken,
        description: form.description,
      });
      Message.success(editMode ? '连接已更新' : '连接已创建');
      onClose();
    } catch {
      Message.error(editMode ? '更新失败' : '创建失败');
    } finally {
      setSubmitting(false);
    }
  };

  if (!visible) return null;

  return (
    <ZModal
      title={editMode ? '编辑 Git 连接' : '添加 Git 连接'}
      onClose={onClose}
      footer={
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
          <Btn onClick={onClose}>取消</Btn>
          <Btn variant="primary" onClick={handleOk} disabled={submitting}>
            {submitting ? '保存中...' : '保存'}
          </Btn>
        </div>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        <Field label="连接名称" required>
          <input className="input" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="例如: my-gitlab" />
        </Field>
        {!editMode && (
          <>
            <Field label="Provider 类型" required>
              <ZSelect
                width={200}
                value={form.providerType}
                options={[{ value: 'gitlab', label: 'GitLab' }, { value: 'github', label: 'GitHub' }]}
                onChange={(v) => setForm({ ...form, providerType: v })}
              />
            </Field>
            <Field
              label="Server URL"
              required
              help="支持内网地址，如 http://192.168.1.100:8080 或 https://git.internal.company.com"
            >
              <input
                className="input mono"
                value={form.serverUrl}
                onChange={(e) => setForm({ ...form, serverUrl: e.target.value })}
                placeholder="https://gitlab.example.com"
              />
            </Field>
          </>
        )}
        <Field
          label="Access Token (PAT)"
          required={!editMode}
          help={editMode ? '留空则不更新' : undefined}
        >
          <input
            className="input mono"
            type="password"
            value={form.accessToken}
            onChange={(e) => setForm({ ...form, accessToken: e.target.value })}
            placeholder={editMode ? '留空则不更新' : 'Personal Access Token'}
          />
          {!editMode && (
            <div style={{ fontSize: 11, color: 'var(--z-500)', marginTop: 6, lineHeight: 1.7 }}>
              <div style={{ fontWeight: 500, color: 'var(--z-700)', marginBottom: 2 }}>如何获取 PAT：</div>
              <div>• <b>GitHub</b>：Settings → Developer settings → Personal access tokens，勾选 <code className="code">repo</code> 权限</div>
              <div>• <b>GitLab</b>：Settings → Access Tokens，勾选 <code className="code">api</code> + <code className="code">read_repository</code></div>
            </div>
          )}
        </Field>
        <Field label="描述">
          <textarea className="input" rows={2} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} placeholder="可选的描述信息" style={{ resize: 'vertical' }} />
        </Field>
      </div>
    </ZModal>
  );
}
