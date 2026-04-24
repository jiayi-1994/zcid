import { useState, useEffect, useCallback, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { Message } from '@arco-design/web-react';
import { AppLayout } from '../../components/layout/AppLayout';
import { useAuthStore } from '../../stores/auth';
import { fetchProjects, deleteProject, type Project } from '../../services/project';
import { ProjectFormModal } from './ProjectFormModal';
import { PageHeader } from '../../components/ui/PageHeader';
import { Metric } from '../../components/ui/Metric';
import { Btn } from '../../components/ui/Btn';
import { StatusBadge } from '../../components/ui/StatusBadge';
import { Avatar } from '../../components/ui/Avatar';
import { IFolder, ICheck, ISearch, IPlus, ITrash, IArrR } from '../../components/ui/icons';

const PROJECT_PAGE_SIZE = 12;

export function ProjectListPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const navigate = useNavigate();
  const isAdmin = useAuthStore((s) => s.user?.role === 'admin');

  const loadProjects = useCallback(async (p: number) => {
    setLoading(true);
    try {
      const data = await fetchProjects(p, PROJECT_PAGE_SIZE);
      setProjects(data.items ?? []);
      setTotal(data.total);
    } catch { Message.error('加载项目列表失败'); }
    finally { setLoading(false); }
  }, []);

  useEffect(() => { loadProjects(page); }, [loadProjects, page]);

  const filtered = useMemo(() => {
    if (!search.trim()) return projects;
    const q = search.toLowerCase();
    return projects.filter((p) => p.name.toLowerCase().includes(q) || p.description?.toLowerCase().includes(q));
  }, [projects, search]);

  const activeCount = projects.filter((p) => p.status === 'active').length;

  const handleDelete = async (id: string) => {
    try { await deleteProject(id); Message.success('项目已删除'); loadProjects(page); }
    catch { Message.error('删除失败'); }
  };

  return (
    <AppLayout>
      <PageHeader
        crumb="Project Directory"
        title="项目管理"
        sub="管理所有项目及其 CI/CD 配置。"
        actions={
          <>
            <div className="input-wrap">
              <ISearch size={13} />
              <input
                className="input input--with-icon"
                style={{ width: 220 }}
                placeholder="搜索项目..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>
            {isAdmin && (
              <Btn size="sm" variant="primary" icon={<IPlus size={13} />} onClick={() => setModalVisible(true)}>
                新建项目
              </Btn>
            )}
          </>
        }
      />

      <div style={{ padding: 24, display: 'flex', flexDirection: 'column', gap: 18 }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 14, maxWidth: 520 }}>
          <Metric label="TOTAL PROJECTS" value={total} icon={<IFolder size={14} />} iconBg="var(--accent-soft)" iconColor="var(--accent-ink)" />
          <Metric label="ACTIVE" value={activeCount} icon={<ICheck size={14} />} iconBg="var(--green-soft)" iconColor="var(--green-ink)" />
        </div>

        {loading ? (
          <div style={{ color: 'var(--z-400)', padding: '40px 0', textAlign: 'center' }}>加载中...</div>
        ) : filtered.length === 0 ? (
          <div style={{ padding: '48px 0', textAlign: 'center', color: 'var(--z-500)' }}>
            <div style={{ fontSize: 14, fontWeight: 500, color: 'var(--z-800)', marginBottom: 4 }}>
              {search ? '没有找到匹配的项目' : '还没有项目'}
            </div>
            <div style={{ fontSize: 12.5, marginBottom: 14 }}>
              {search ? '尝试调整搜索关键词' : '创建你的第一个项目，开始 CI/CD 之旅'}
            </div>
            {!search && isAdmin && (
              <Btn variant="primary" icon={<IPlus size={13} />} onClick={() => setModalVisible(true)}>
                创建第一个项目
              </Btn>
            )}
          </div>
        ) : (
          <>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: 14 }}>
              {filtered.map((p) => {
                const hue = (p.name.charCodeAt(0) * 17) % 360;
                return (
                  <div
                    key={p.id}
                    className="card"
                    style={{ padding: 0, overflow: 'hidden', cursor: 'pointer' }}
                    onClick={() => navigate(`/projects/${p.id}/pipelines`)}
                  >
                    <div style={{ padding: '14px 14px 10px', display: 'flex', gap: 10, alignItems: 'flex-start' }}>
                      <div style={{
                        width: 36, height: 36, borderRadius: 8, flexShrink: 0,
                        background: `linear-gradient(135deg, oklch(0.64 0.17 ${hue}), oklch(0.52 0.19 ${(hue + 24) % 360}))`,
                        color: '#fff', display: 'flex', alignItems: 'center', justifyContent: 'center',
                        fontWeight: 700, fontSize: 15,
                      }}>
                        {p.name.charAt(0).toUpperCase()}
                      </div>
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={{ fontSize: 13.5, fontWeight: 600 }}>{p.name}</div>
                        <StatusBadge status={p.status === 'active' ? 'active' : 'userDisabled'} />
                      </div>
                    </div>
                    <div style={{ padding: '0 14px 12px', fontSize: 12, color: 'var(--z-600)', minHeight: 34, overflow: 'hidden', display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical' as const }}>
                      {p.description || '暂无项目描述'}
                    </div>
                    <div style={{ padding: '10px 14px', borderTop: '1px solid var(--z-150)', background: 'var(--z-25)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                      <span className="sub" style={{ fontSize: 11 }}>{new Date(p.createdAt).toLocaleDateString()}</span>
                      <div style={{ display: 'flex', gap: 6 }} onClick={(e) => e.stopPropagation()}>
                        {isAdmin && (
                          <Btn size="xs" variant="ghost" iconOnly icon={<ITrash size={12} />} onClick={() => handleDelete(p.id)} />
                        )}
                        <Btn size="xs" variant="outline" onClick={() => navigate(`/projects/${p.id}/pipelines`)}>
                          进入 <IArrR size={11} />
                        </Btn>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>

            {total > PROJECT_PAGE_SIZE && (
              <div style={{ display: 'flex', gap: 4, justifyContent: 'flex-end', fontSize: 11.5, color: 'var(--z-500)' }}>
                <span style={{ paddingRight: 8, lineHeight: '22px' }}>共 {total} 个</span>
                <Btn size="xs" variant="ghost" disabled={page === 1} onClick={() => setPage(page - 1)}>上一页</Btn>
                <Btn size="xs" variant="outline">{page}</Btn>
                <Btn size="xs" variant="ghost" disabled={page * PROJECT_PAGE_SIZE >= total} onClick={() => setPage(page + 1)}>下一页</Btn>
              </div>
            )}
          </>
        )}
      </div>

      <ProjectFormModal
        visible={modalVisible}
        onClose={() => setModalVisible(false)}
        onSuccess={() => { setModalVisible(false); loadProjects(page); }}
      />
    </AppLayout>
  );
}
