import { Button, Table, Space, Message, Popconfirm, Tag } from '@arco-design/web-react';
import { IconPlus } from '@arco-design/web-react/icon';
import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { AppLayout } from '../../components/layout/AppLayout';
import { useAuthStore } from '../../stores/auth';
import { fetchProjects, deleteProject, type Project } from '../../services/project';
import { ProjectFormModal } from './ProjectFormModal';

export function ProjectListPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const navigate = useNavigate();
  const isAdmin = useAuthStore((s) => s.user?.role === 'admin');

  const loadProjects = useCallback(async (p: number) => {
    setLoading(true);
    try {
      const data = await fetchProjects(p, 20);
      setProjects(data.items ?? []);
      setTotal(data.total);
    } catch (err: any) {
      Message.error(err.response?.data?.message || '加载项目列表失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadProjects(page);
  }, [page, loadProjects]);

  const handleDelete = async (id: string) => {
    try {
      await deleteProject(id);
      Message.success('项目已删除');
      loadProjects(page);
    } catch (err: any) {
      Message.error(err.response?.data?.message || '删除失败');
    }
  };

  const columns = [
    {
      title: '项目名称',
      dataIndex: 'name',
      render: (name: string, record: Project) => (
        <Button type="text" onClick={() => navigate(`/projects/${record.id}/environments`)}
          style={{ padding: 0, fontWeight: 500, color: 'var(--zcid-text-1)' }}
        >
          {name}
        </Button>
      ),
    },
    { title: '描述', dataIndex: 'description', render: (v: string) => <span style={{ color: 'var(--zcid-text-3)' }}>{v || '-'}</span> },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (status: string) => (
        <Tag size="small" color={status === 'active' ? 'green' : 'default'}>{status}</Tag>
      ),
    },
    { title: '创建时间', dataIndex: 'createdAt', width: 180 },
    {
      title: '操作',
      width: 140,
      render: (_: any, record: Project) => (
        <Space size="mini">
          <Button type="text" size="small" onClick={() => navigate(`/projects/${record.id}/environments`)}
            style={{ color: 'var(--zcid-primary)' }}
          >
            进入
          </Button>
          {isAdmin && (
            <Popconfirm title="确定删除此项目？" onOk={() => handleDelete(record.id)}>
              <Button type="text" size="small" status="danger">删除</Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  return (
    <AppLayout>
      <div className="page-container">
        <div className="page-header">
          <div>
            <h3 className="page-title">项目管理</h3>
            <p className="page-subtitle">管理所有项目及其 CI/CD 配置</p>
          </div>
          {isAdmin && (
            <Button type="primary" icon={<IconPlus />} onClick={() => setModalVisible(true)}>
              新建项目
            </Button>
          )}
        </div>
        <div className="table-card">
          <Table
            columns={columns}
            data={projects}
            loading={loading}
            rowKey="id"
            border={false}
            pagination={{
              current: page,
              total,
              pageSize: 20,
              onChange: setPage,
              style: { padding: '12px 16px' },
            }}
            noDataElement={
              <div className="empty-state">
                <div className="empty-state-title">暂无项目</div>
                <div className="empty-state-desc">创建你的第一个项目，开始 CI/CD 之旅</div>
              </div>
            }
          />
        </div>
        <ProjectFormModal
          visible={modalVisible}
          onClose={() => setModalVisible(false)}
          onSuccess={() => { setModalVisible(false); loadProjects(page); }}
        />
      </div>
    </AppLayout>
  );
}
