import { useEffect, useState } from 'react';
import { Message } from '@arco-design/web-react';
import { createProject } from '../../services/project';
import { extractErrorMessage } from '../../services/http';
import { ZModal } from '../../components/ui/ZModal';
import { Btn } from '../../components/ui/Btn';
import { Field } from '../../components/ui/Field';

interface Props {
  visible: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export function ProjectFormModal({ visible, onClose, onSuccess }: Props) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!visible) { setName(''); setDescription(''); }
  }, [visible]);

  const handleSubmit = async () => {
    if (!name.trim()) { Message.error('请输入项目名称'); return; }
    setLoading(true);
    try {
      await createProject(name.trim(), description);
      Message.success('项目创建成功');
      onSuccess();
    } catch (err: unknown) {
      Message.error(extractErrorMessage(err, '创建失败'));
    } finally {
      setLoading(false);
    }
  };

  if (!visible) return null;

  return (
    <ZModal
      title="新建项目"
      onClose={onClose}
      footer={
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
          <Btn onClick={onClose}>取消</Btn>
          <Btn variant="primary" onClick={handleSubmit} disabled={loading}>
            {loading ? '创建中...' : '创建'}
          </Btn>
        </div>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        <Field label="项目名称" required>
          <input className="input" value={name} onChange={(e) => setName(e.target.value)} placeholder="输入项目名称" autoFocus />
        </Field>
        <Field label="描述">
          <textarea className="input" rows={3} value={description} onChange={(e) => setDescription(e.target.value)} placeholder="输入项目描述（可选）" style={{ resize: 'vertical' }} />
        </Field>
      </div>
    </ZModal>
  );
}
