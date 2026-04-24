import { useCallback, useEffect, useState } from 'react';
import { Message } from '@arco-design/web-react';
import { useParams } from 'react-router-dom';
import { useAuthStore } from '../../../stores/auth';
import { fetchServices, createService, deleteService, type ServiceItem } from '../../../services/project';
import { extractErrorMessage } from '../../../services/http';
import { PageHeader } from '../../../components/ui/PageHeader';
import { Card } from '../../../components/ui/Card';
import { Btn } from '../../../components/ui/Btn';
import { ZModal } from '../../../components/ui/ZModal';
import { Field } from '../../../components/ui/Field';
import { IPlus, IChevL, IChevR } from '../../../components/ui/icons';

export function ServiceListPage() {
  const { id: projectId } = useParams<{ id: string }>();
  const [svcs, setSvcs] = useState<ServiceItem[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [form, setForm] = useState({ name: '', description: '', repoUrl: '' });
  const isAdmin = useAuthStore((s) => s.user?.role === 'admin');

  const loadData = useCallback(async (p: number) => {
    if (!projectId) return;
    setLoading(true);
    try {
      const data = await fetchServices(projectId, p, 20);
      setSvcs(data.items ?? []);
      setTotal(data.total);
    } catch (err: unknown) {
      Message.error(extractErrorMessage(err, '加载服务列表失败'));
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => { loadData(page); }, [page, loadData]);

  const handleCreate = async () => {
    if (!form.name) { Message.error('请输入服务名称'); return; }
    setSubmitting(true);
    try {
      await createService(projectId!, form.name, form.description, form.repoUrl);
      Message.success('服务创建成功');
      setForm({ name: '', description: '', repoUrl: '' });
      setModalVisible(false);
      loadData(page);
    } catch (err: unknown) {
      Message.error(extractErrorMessage(err, '创建失败'));
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (svcId: string) => {
    try {
      await deleteService(projectId!, svcId);
      Message.success('服务已删除');
      loadData(page);
    } catch (err: unknown) {
      Message.error(extractErrorMessage(err, '删除失败'));
    }
  };

  const totalPages = Math.ceil(total / 20);

  return (
    <>
      <PageHeader
        crumb="Project · Services"
        title="服务管理"
        sub="管理项目下的微服务与源码仓库绑定。"
        actions={isAdmin && (
          <Btn size="sm" variant="primary" icon={<IPlus size={13} />} onClick={() => setModalVisible(true)}>
            新建服务
          </Btn>
        )}
      />
      <div style={{ padding: 24, display: 'flex', flexDirection: 'column', gap: 14 }}>
        <Card padding={false}>
          {loading ? (
            <div style={{ padding: '40px 0', textAlign: 'center', color: 'var(--z-400)' }}>加载中...</div>
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>服务名称</th><th>描述</th><th>仓库地址</th><th>创建时间</th>
                  {isAdmin && <th style={{ textAlign: 'right' }}>操作</th>}
                </tr>
              </thead>
              <tbody>
                {svcs.map((svc) => (
                  <tr key={svc.id}>
                    <td><span style={{ fontWeight: 500 }}>{svc.name}</span></td>
                    <td><span className="sub">{svc.description || '-'}</span></td>
                    <td>
                      {svc.repoUrl
                        ? <span className="code" style={{ fontSize: 11.5 }}>{svc.repoUrl}</span>
                        : <span className="sub">-</span>}
                    </td>
                    <td><span className="sub mono" style={{ fontSize: 11.5 }}>{svc.createdAt}</span></td>
                    {isAdmin && (
                      <td style={{ textAlign: 'right' }}>
                        <Btn size="xs" variant="ghost" onClick={() => handleDelete(svc.id)}>删除</Btn>
                      </td>
                    )}
                  </tr>
                ))}
                {svcs.length === 0 && !loading && (
                  <tr>
                    <td colSpan={isAdmin ? 5 : 4} style={{ textAlign: 'center', padding: '40px 0', color: 'var(--z-400)' }}>
                      暂无服务
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          )}
        </Card>

        {totalPages > 1 && (
          <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 11.5, color: 'var(--z-500)' }}>
            <span>共 {total} 条 · 第 {page} / {totalPages} 页</span>
            <div style={{ display: 'flex', gap: 4 }}>
              <Btn size="xs" variant="ghost" iconOnly icon={<IChevL size={12} />} disabled={page <= 1} onClick={() => setPage((p) => p - 1)} />
              <Btn size="xs" variant="outline">{page}</Btn>
              <Btn size="xs" variant="ghost" iconOnly icon={<IChevR size={12} />} disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)} />
            </div>
          </div>
        )}
      </div>

      {modalVisible && (
        <ZModal
          title="新建服务"
          onClose={() => setModalVisible(false)}
          footer={
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
              <Btn onClick={() => setModalVisible(false)}>取消</Btn>
              <Btn variant="primary" onClick={handleCreate} disabled={submitting}>
                {submitting ? '创建中...' : '创建'}
              </Btn>
            </div>
          }
        >
          <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <Field label="服务名称" required>
              <input className="input" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="如 api-gateway" />
            </Field>
            <Field label="描述">
              <textarea className="input" rows={2} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} placeholder="服务描述" style={{ resize: 'vertical' }} />
            </Field>
            <Field label="仓库地址">
              <input className="input" value={form.repoUrl} onChange={(e) => setForm({ ...form, repoUrl: e.target.value })} placeholder="https://github.com/org/repo" />
            </Field>
          </div>
        </ZModal>
      )}
    </>
  );
}
