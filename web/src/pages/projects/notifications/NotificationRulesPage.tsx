import { useCallback, useEffect, useState } from 'react';
import { Message } from '@arco-design/web-react';
import { useParams } from 'react-router-dom';
import {
  createNotificationRule, deleteNotificationRule, fetchNotificationRules, updateNotificationRule,
  type NotificationRule,
} from '../../../services/notification';
import { PageHeader } from '../../../components/ui/PageHeader';
import { Card } from '../../../components/ui/Card';
import { Btn } from '../../../components/ui/Btn';
import { Badge } from '../../../components/ui/Badge';
import { ZSwitch } from '../../../components/ui/ZSwitch';
import { ZSelect } from '../../../components/ui/ZSelect';
import { ZModal } from '../../../components/ui/ZModal';
import { Field } from '../../../components/ui/Field';
import { IPlus, IEdit, ITrash } from '../../../components/ui/icons';

const EVENT_OPTIONS = [
  { value: 'build_success', label: '构建成功' },
  { value: 'build_failed', label: '构建失败' },
  { value: 'deploy_success', label: '部署成功' },
  { value: 'deploy_failed', label: '部署失败' },
];

const EVENT_TONE: Record<string, 'green' | 'red' | 'blue' | 'amber'> = {
  build_success: 'green',
  build_failed: 'red',
  deploy_success: 'blue',
  deploy_failed: 'amber',
};

function eventLabel(val: string) {
  return EVENT_OPTIONS.find((o) => o.value === val)?.label ?? val;
}

const EMPTY_FORM = { name: '', eventType: 'build_failed', webhookUrl: '', enabled: true };

export function NotificationRulesPage() {
  const { id: projectId } = useParams<{ id: string }>();
  const [rules, setRules] = useState<NotificationRule[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editRule, setEditRule] = useState<NotificationRule | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [form, setForm] = useState(EMPTY_FORM);

  const load = useCallback(async () => {
    if (!projectId) return;
    setLoading(true);
    try {
      const data = await fetchNotificationRules(projectId);
      setRules(data.items || []);
    } catch {
      Message.error('加载通知规则失败');
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => { load(); }, [load]);

  const openCreate = () => {
    setEditRule(null);
    setForm(EMPTY_FORM);
    setModalVisible(true);
  };

  const openEdit = (rule: NotificationRule) => {
    setEditRule(rule);
    setForm({ name: rule.name, eventType: rule.eventType, webhookUrl: rule.webhookUrl, enabled: rule.enabled });
    setModalVisible(true);
  };

  const handleSubmit = async () => {
    if (!form.name || !form.webhookUrl) { Message.error('请填写名称和 Webhook URL'); return; }
    if (!projectId) return;
    setSubmitting(true);
    try {
      if (editRule) {
        await updateNotificationRule(projectId, editRule.id, form);
        Message.success('通知规则已更新');
      } else {
        await createNotificationRule(projectId, form);
        Message.success('通知规则已创建');
      }
      setModalVisible(false);
      load();
    } catch {
      Message.error('操作失败');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (ruleId: string) => {
    if (!projectId) return;
    try {
      await deleteNotificationRule(projectId, ruleId);
      Message.success('通知规则已删除');
      load();
    } catch {
      Message.error('删除失败');
    }
  };

  const handleToggle = async (rule: NotificationRule, enabled: boolean) => {
    if (!projectId) return;
    try {
      await updateNotificationRule(projectId, rule.id, { enabled });
      load();
    } catch {
      Message.error('更新状态失败');
    }
  };

  return (
    <>
      <PageHeader
        crumb="Project · Signals"
        title="通知规则"
        sub="配置构建与部署事件的 Webhook 推送。"
        actions={
          <Btn size="sm" variant="primary" icon={<IPlus size={13} />} onClick={openCreate}>
            创建规则
          </Btn>
        }
      />
      <div style={{ padding: 24 }}>
        <Card padding={false}>
          {loading ? (
            <div style={{ padding: '40px 0', textAlign: 'center', color: 'var(--z-400)' }}>加载中...</div>
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>名称</th><th>事件类型</th><th>Webhook URL</th><th>状态</th><th>创建时间</th>
                  <th style={{ textAlign: 'right' }}>操作</th>
                </tr>
              </thead>
              <tbody>
                {rules.map((rule) => (
                  <tr key={rule.id}>
                    <td><span style={{ fontWeight: 500 }}>{rule.name}</span></td>
                    <td>
                      <Badge tone={EVENT_TONE[rule.eventType] ?? 'blue'}>
                        {eventLabel(rule.eventType)}
                      </Badge>
                    </td>
                    <td>
                      <span className="code" style={{ fontSize: 11, maxWidth: 260, display: 'inline-block', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', verticalAlign: 'middle' }}>
                        {rule.webhookUrl}
                      </span>
                    </td>
                    <td>
                      <ZSwitch on={rule.enabled} onChange={(v) => handleToggle(rule, v)} />
                    </td>
                    <td><span className="sub mono" style={{ fontSize: 11 }}>{rule.createdAt}</span></td>
                    <td style={{ textAlign: 'right' }}>
                      <div style={{ display: 'inline-flex', gap: 4 }}>
                        <Btn size="xs" variant="ghost" iconOnly icon={<IEdit size={12} />} onClick={() => openEdit(rule)} />
                        <Btn size="xs" variant="ghost" iconOnly icon={<ITrash size={12} />} onClick={() => handleDelete(rule.id)} />
                      </div>
                    </td>
                  </tr>
                ))}
                {rules.length === 0 && !loading && (
                  <tr>
                    <td colSpan={6} style={{ textAlign: 'center', padding: '40px 0', color: 'var(--z-400)' }}>
                      暂无通知规则
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          )}
        </Card>
      </div>

      {modalVisible && (
        <ZModal
          title={editRule ? '编辑通知规则' : '创建通知规则'}
          onClose={() => setModalVisible(false)}
          footer={
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
              <Btn onClick={() => setModalVisible(false)}>取消</Btn>
              <Btn variant="primary" onClick={handleSubmit} disabled={submitting}>
                {submitting ? '保存中...' : '保存'}
              </Btn>
            </div>
          }
        >
          <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <Field label="名称" required>
              <input className="input" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="例如：构建失败通知" />
            </Field>
            <Field label="事件类型" required>
              <ZSelect
                width={220}
                value={form.eventType}
                options={EVENT_OPTIONS}
                onChange={(v) => setForm({ ...form, eventType: v })}
              />
            </Field>
            <Field label="Webhook URL" required>
              <input className="input" value={form.webhookUrl} onChange={(e) => setForm({ ...form, webhookUrl: e.target.value })} placeholder="https://hooks.example.com/callback" />
            </Field>
            <Field label="启用">
              <ZSwitch on={form.enabled} onChange={(v) => setForm({ ...form, enabled: v })} />
            </Field>
          </div>
        </ZModal>
      )}
    </>
  );
}

export default NotificationRulesPage;
