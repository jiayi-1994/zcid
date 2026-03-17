import { Button, Grid, Message, Popconfirm, Tag, Skeleton, Input } from '@arco-design/web-react';
import { IconPlus, IconSearch, IconThunderbolt, IconClockCircle, IconArrowRight } from '@arco-design/web-react/icon';
import { useState, useEffect, useCallback, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { AppLayout } from '../../components/layout/AppLayout';
import { useAuthStore } from '../../stores/auth';
import { fetchProjects, deleteProject, type Project } from '../../services/project';
import { ProjectFormModal } from './ProjectFormModal';

const { Row, Col } = Grid;

const STATUS_CONFIG: Record<string, { color: string; bg: string; label: string }> = {
  active: { color: '#00B42A', bg: '#E8FFEA', label: '运行中' },
  inactive: { color: '#86909C', bg: '#F2F3F5', label: '未激活' },
  archived: { color: '#FF7D00', bg: '#FFF7E8', label: '已归档' },
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
    <div
      style={{
        background: '#fff',
        borderRadius: 12,
        border: '1px solid var(--zcid-border)',
        padding: '20px',
        cursor: 'pointer',
        transition: 'all 0.2s',
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
      }}
      onClick={() => onEnter(project.id)}
      onMouseEnter={(e) => {
        (e.currentTarget as HTMLElement).style.boxShadow = '0 4px 16px rgba(0,0,0,0.08)';
        (e.currentTarget as HTMLElement).style.borderColor = '#165DFF';
        (e.currentTarget as HTMLElement).style.transform = 'translateY(-2px)';
      }}
      onMouseLeave={(e) => {
        (e.currentTarget as HTMLElement).style.boxShadow = 'none';
        (e.currentTarget as HTMLElement).style.borderColor = 'var(--zcid-border)';
        (e.currentTarget as HTMLElement).style.transform = 'none';
      }}
    >
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 12 }}>
        <div style={{
          width: 40, height: 40, borderRadius: 10, flexShrink: 0,
          background: `linear-gradient(135deg, #165DFF 0%, #0FC6C2 100%)`,
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          fontSize: 16, fontWeight: 700, color: '#fff',
          boxShadow: '0 2px 8px rgba(22,93,255,0.2)',
        }}>
          {initial}
        </div>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ fontSize: 15, fontWeight: 600, color: '#1D2129', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {project.name}
          </div>
          <span style={{
            display: 'inline-flex', alignItems: 'center', gap: 4,
            padding: '1px 8px', borderRadius: 10, fontSize: 11, fontWeight: 500,
            background: cfg.bg, color: cfg.color,
            marginTop: 2,
          }}>
            <span style={{ width: 5, height: 5, borderRadius: '50%', background: cfg.color }} />
            {cfg.label}
          </span>
        </div>
      </div>

      {/* Description */}
      <div style={{
        fontSize: 13, color: '#86909C', lineHeight: 1.5,
        flex: 1, marginBottom: 12,
        overflow: 'hidden', textOverflow: 'ellipsis',
        display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical' as const,
      }}>
        {project.description || '暂无项目描述'}
      </div>

      {/* Footer */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', paddingTop: 12, borderTop: '1px solid #F2F3F5' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 4, fontSize: 12, color: '#C9CDD4' }}>
          <IconClockCircle style={{ fontSize: 12 }} />
          {new Date(project.createdAt).toLocaleDateString()}
        </div>
        <div style={{ display: 'flex', gap: 4 }} onClick={(e) => e.stopPropagation()}>
          {canDelete && (
            <Popconfirm title="确定删除此项目？" onOk={() => onDelete(project.id)}>
              <Button size="mini" type="text" status="danger" style={{ borderRadius: 6, fontSize: 12 }}>删除</Button>
            </Popconfirm>
          )}
          <Button size="mini" type="outline" icon={<IconArrowRight />} onClick={() => onEnter(project.id)} style={{ borderRadius: 6, fontSize: 12 }}>
            进入
          </Button>
        </div>
      </div>
    </div>
  );
}

export function ProjectListPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [search, setSearch] = useState('');
  const navigate = useNavigate();
  const isAdmin = useAuthStore((s) => s.user?.role === 'admin');

  const loadProjects = useCallback(async () => {
    setLoading(true);
    try {
      const data = await fetchProjects(1, 100);
      setProjects(data.items ?? []);
    } catch {
      Message.error('加载项目列表失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { loadProjects(); }, [loadProjects]);

  const filteredProjects = useMemo(() => {
    if (!search.trim()) return projects;
    const q = search.toLowerCase();
    return projects.filter((p) => p.name.toLowerCase().includes(q) || p.description?.toLowerCase().includes(q));
  }, [projects, search]);

  const handleDelete = async (id: string) => {
    try {
      await deleteProject(id);
      Message.success('项目已删除');
      loadProjects();
    } catch {
      Message.error('删除失败');
    }
  };

  return (
    <AppLayout>
      <div className="page-container">
        <div className="page-header">
          <div>
            <h3 className="page-title">项目管理</h3>
            <p className="page-subtitle">管理所有项目及其 CI/CD 配置</p>
          </div>
          <div style={{ display: 'flex', gap: 12 }}>
            <Input
              prefix={<IconSearch />}
              placeholder="搜索项目..."
              value={search}
              onChange={setSearch}
              allowClear
              style={{ width: 220, borderRadius: 8 }}
            />
            {isAdmin && (
              <Button type="primary" icon={<IconPlus />} onClick={() => setModalVisible(true)} style={{ borderRadius: 8 }}>
                新建项目
              </Button>
            )}
          </div>
        </div>

        {/* Stats bar */}
        <div style={{
          display: 'flex', gap: 24, marginBottom: 20,
          padding: '12px 20px', background: '#F7F8FA', borderRadius: 10,
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <IconThunderbolt style={{ fontSize: 16, color: '#165DFF' }} />
            <span style={{ fontSize: 13, color: '#4E5969' }}>共 <strong style={{ color: '#1D2129' }}>{projects.length}</strong> 个项目</span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <span style={{ width: 6, height: 6, borderRadius: '50%', background: '#00B42A' }} />
            <span style={{ fontSize: 13, color: '#4E5969' }}><strong style={{ color: '#1D2129' }}>{projects.filter(p => p.status === 'active').length}</strong> 个活跃</span>
          </div>
        </div>

        {loading ? (
          <Row gutter={[16, 16]}>
            {[1, 2, 3, 4].map((i) => (
              <Col key={i} xs={24} sm={12} md={8} lg={6}>
                <div style={{ background: '#fff', borderRadius: 12, border: '1px solid var(--zcid-border)', padding: 20 }}>
                  <Skeleton text={{ rows: 3 }} animation />
                </div>
              </Col>
            ))}
          </Row>
        ) : filteredProjects.length === 0 ? (
          <div style={{
            textAlign: 'center', padding: '80px 24px',
            background: '#fff', borderRadius: 12, border: '1px solid var(--zcid-border)',
          }}>
            <div style={{ fontSize: 48, marginBottom: 16 }}>📦</div>
            <div style={{ fontSize: 16, fontWeight: 600, color: '#1D2129', marginBottom: 8 }}>
              {search ? '没有找到匹配的项目' : '还没有项目'}
            </div>
            <div style={{ fontSize: 14, color: '#86909C', marginBottom: 20 }}>
              {search ? '尝试调整搜索关键词' : '创建你的第一个项目，开始 CI/CD 之旅'}
            </div>
            {!search && isAdmin && (
              <Button type="primary" icon={<IconPlus />} onClick={() => setModalVisible(true)} style={{ borderRadius: 8 }}>
                创建第一个项目
              </Button>
            )}
          </div>
        ) : (
          <Row gutter={[16, 16]}>
            {filteredProjects.map((p) => (
              <Col key={p.id} xs={24} sm={12} md={8} lg={6}>
                <ProjectCard
                  project={p}
                  onEnter={(id) => navigate(`/projects/${id}/pipelines`)}
                  onDelete={handleDelete}
                  canDelete={isAdmin}
                />
              </Col>
            ))}
          </Row>
        )}

        <ProjectFormModal
          visible={modalVisible}
          onClose={() => setModalVisible(false)}
          onSuccess={() => { setModalVisible(false); loadProjects(); }}
        />
      </div>
    </AppLayout>
  );
}
