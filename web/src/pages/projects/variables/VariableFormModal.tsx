import { useEffect, useState } from 'react';
import { Message } from '@arco-design/web-react';
import { ZModal } from '../../../components/ui/ZModal';
import { Btn } from '../../../components/ui/Btn';
import { Field } from '../../../components/ui/Field';

interface VariableFormModalProps {
  visible: boolean;
  onClose: () => void;
  onSubmit: (data: { key: string; value: string; varType: string; description: string }) => Promise<void>;
  editMode?: boolean;
  isSecret?: boolean;
  initialValues?: { key?: string; description?: string };
}

export function VariableFormModal({ visible, onClose, onSubmit, editMode, isSecret, initialValues }: VariableFormModalProps) {
  const [form, setForm] = useState({ key: '', value: '', varType: 'plain', description: '' });
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (visible) {
      setForm({
        key: initialValues?.key ?? '',
        value: '',
        varType: isSecret ? 'secret' : 'plain',
        description: initialValues?.description ?? '',
      });
    }
  }, [visible, initialValues, isSecret]);

  const handleOk = async () => {
    if (!editMode && !form.key.trim()) { Message.error('请输入变量名'); return; }
    if (!editMode && !form.value) { Message.error('请输入变量值'); return; }
    setLoading(true);
    try {
      await onSubmit({
        key: form.key.trim(),
        value: form.value,
        varType: form.varType,
        description: form.description,
      });
      Message.success(editMode ? '变量更新成功' : '变量创建成功');
      onClose();
    } catch {
      /* surfaced by caller */
    } finally {
      setLoading(false);
    }
  };

  if (!visible) return null;

  return (
    <ZModal
      title={editMode ? '编辑变量' : '新建变量'}
      onClose={onClose}
      footer={
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
          <Btn onClick={onClose}>取消</Btn>
          <Btn variant="primary" onClick={handleOk} disabled={loading}>
            {loading ? '保存中...' : '保存'}
          </Btn>
        </div>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        {!editMode && (
          <Field label="变量名" required>
            <input
              className="input mono"
              value={form.key}
              onChange={(e) => setForm({ ...form, key: e.target.value })}
              placeholder="例如: DB_HOST"
            />
          </Field>
        )}
        <Field label="值" required={!editMode}>
          <input
            className="input mono"
            type={isSecret ? 'password' : 'text'}
            value={form.value}
            onChange={(e) => setForm({ ...form, value: e.target.value })}
            placeholder={isSecret ? '输入新的密钥值' : '变量值'}
          />
        </Field>
        {!editMode && (
          <Field label="类型">
            <div style={{ display: 'flex', gap: 12 }}>
              {[{ value: 'plain', label: '普通' }, { value: 'secret', label: '密钥' }].map((opt) => (
                <label key={opt.value} style={{ display: 'inline-flex', alignItems: 'center', gap: 6, cursor: 'pointer', fontSize: 13 }}>
                  <input
                    type="radio"
                    name="varType"
                    value={opt.value}
                    checked={form.varType === opt.value}
                    onChange={() => setForm({ ...form, varType: opt.value })}
                  />
                  {opt.label}
                </label>
              ))}
            </div>
          </Field>
        )}
        <Field label="描述">
          <textarea
            className="input"
            rows={2}
            value={form.description}
            onChange={(e) => setForm({ ...form, description: e.target.value })}
            placeholder="变量描述（可选）"
            style={{ resize: 'vertical' }}
          />
        </Field>
      </div>
    </ZModal>
  );
}
