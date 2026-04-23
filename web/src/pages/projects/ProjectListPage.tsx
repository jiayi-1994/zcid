import { Button, Message, Popconfirm, Skeleton, Input, Pagination } from '@arco-design/web-react';
import {
  IconPlus,
  IconSearch,
  IconApps,
  IconCheckCircle,
  IconClockCircle,
  IconArrowRight,
} from '@arco-design/web-react/icon';
import { useState, useEffect, useCallback, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { AppLayout } from '../../components/layout/AppLayout';
import { useAuthStore } from '../../stores/auth';
import { fetchProjects, deleteProject, type Project } from '../../services/project';
import { ProjectFormModal } from './ProjectFormModal';

const STATUS_CONFIG: Record<string, { cls: string; label: string }> = {
  active: { cls: 'pipeline-status-badge--success', label: '运行中' },
  inactive: { cls: 'pipeline-status-badge--pending', label: '未激活' },
  archived: { cls: 'pipeline-status-badge--cancelled', label: '已归档' },
};

function ProjectCard({ project, onEnter, onDelete, canDelete }: {
  project: Project;
  onEnter: (id: string) => void;
  onDelete: (id: string) => void;
  canDelete: boolean;
}) {
  const cfg = STATUS_CONFIG[project.status] || STATUS_CONFIG.active;
  const initial = project.name.charAt(0).toUpperCase();

  return (
    <div className="project-card zcid-card-interactive" onClick={() => onEnter(project.id)}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 12 }}>
        <div
          className="stat-card-icon stat-card-icon--primary"
          style={{ background: 'var(--primary-gradient)', color: '#fff' }}
        >
          {initial}
        </div>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div className="pipeline-status-name" style={{ marginBottom: 2 }}>
            {project.name}
          </div>
          <span className={`pipeline-status-badge ${cfg.cls}`}>{cfg.label}</span>
        </div>
      </div>

      <div
        style={{
          fontSize: 13,
          color: 'var(--on-surface-variant)',
          lineHeight: 1.5,
          flex: 1,
          marginBottom: 12,
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          display: '-webkit-box',
          WebkitLineClamp: 2,
          WebkitBoxOrient: 'vertical' as const,
        }}
      >
        {project.description || '暂无项目描述'}
      </div>

      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          paddingTop: 12,
          borderTop: '1px solid var(--ghost-border)',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 4,
            fontSize: 12,
            color: 'var(--on-surface-variant)',
          }}
        >
          <IconClockCircle style={{ fontSize: 12 }} />
          {new Date(project.createdAt).toLocaleDateString()}
        </div>
        <div style={{ display: 'flex', gap: 4 }} onClick={(e) => e.stopPropagation()}>
          {canDelete && (
            <Popconfirm title="确定删除此项目？" onOk={() => onDelete(project.id)}>
              <Button size="mini" type="text" status="danger">
                删除
              </Button>
            </Popconfirm>
          )}
          <Button
            size="mini"
            type="outline"
            icon={<IconArrowRight />}
            onClick={() => onEnter(project.id)}
          >
            进入
          </Button>
        </div>
      </div>
    </div>
  );
}

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
    } catch {
      Message.error('加载项目列表失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadProjects(page);
  }, [loadProjects, page]);

  const filteredProjects = useMemo(() => {
    if (!search.trim()) return projects;
    const q = search.toLowerCase();
    return projects.filter(
      (p) =>
        p.name.toLowerCase().includes(q) ||
        p.description?.toLowerCase().includes(q),
    );
  }, [projects, search]);

  const activeCount = projects.filter((p) => p.status === 'active').length;

  const handleDelete = async (id: string) => {
    try {
      await deleteProject(id);
      Message.success('项目已删除');
      loadProjects(page);
    } catch {
      Message.error('删除失败');
    }
  };

  return (
    <AppLayout>
      <div className="page-container">
        <div className="page-header">
          <div>
            <div className="breadcrumb">Project Directory</div>
            <h1 className="page-title">项目管理</h1>
            <p className="page-subtitle">管理所有项目及其 CI/CD 配置</p>
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            <Input
              prefix={<IconSearch />}
              placeholder="搜索项目..."
              value={search}
              onChange={setSearch}
              allowClear
              style={{ width: 220 }}
            />
            {isAdmin && (
              <Button type="primary" icon={<IconPlus />} onClick={() => setModalVisible(true)}>
                新建项目
              </Button>
            )}
          </div>
        </div>

        <div className="metrics-grid" style={{ gridTemplateColumns: 'repeat(2, 1fr)' }}>
          <div className="metric-card">
            <div className="stat-card-icon stat-card-icon--primary">
              <IconApps />
            </div>
            <div className="metric-card-label">Total Projects</div>
            <div className="metric-card-value">{total}</div>
            <div className="metric-card-sub">全量项目数</div>
          </div>
          <div className="metric-card">
            <div className="stat-card-icon stat-card-icon--success">
              <IconCheckCircle />
            </div>
            <div className="metric-card-label">Active</div>
            <div className="metric-card-value">{activeCount}</div>
            <div className="metric-card-sub">当前活跃</div>
          </div>
        </div>

        {loading ? (
          <div className="metrics-grid">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="project-card">
                <Skeleton text={{ rows: 3 }} animation />
              </div>
            ))}
          </div>
        ) : filteredProjects.length === 0 ? (
          <div className="zcid-card empty-state">
            <div className="empty-state-title">
              {search ? '没有找到匹配的项目' : '还没有项目'}
            </div>
            <div className="empty-state-desc">
              {search ? '尝试调整搜索关键词' : '创建你的第一个项目，开始 CI/CD 之旅'}
            </div>
            {!search && isAdmin && (
              <Button type="primary" icon={<IconPlus />} onClick={() => setModalVisible(true)}>
                创建第一个项目
              </Button>
            )}
          </div>
        ) : (
          <>
            <div className="metrics-grid" style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))' }}>
              {filteredProjects.map((p) => (
                <ProjectCard
                  key={p.id}
                  project={p}
                  onEnter={(id) => navigate(`/projects/${id}/pipelines`)}
                  onDelete={handleDelete}
                  canDelete={isAdmin}
                />
              ))}
            </div>
            {total > PROJECT_PAGE_SIZE && (
              <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 'var(--space-5)' }}>
                <Pagination
                  current={page}
                  pageSize={PROJECT_PAGE_SIZE}
                  total={total}
                  onChange={setPage}
                  showTotal
                />
              </div>
            )}
          </>
        )}

        <ProjectFormModal
          visible={modalVisible}
          onClose={() => setModalVisible(false)}
          onSuccess={() => {
            setModalVisible(false);
            loadProjects(page);
          }}
        />
      </div>
    </AppLayout>
  );
}
