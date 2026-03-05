import { Button, Card, Space, Typography } from '@arco-design/web-react';
import { IconApps, IconLink, IconPlayArrow } from '@arco-design/web-react/icon';
import { useNavigate } from 'react-router-dom';

export const ONBOARDING_DISMISSED_KEY = 'zcid_onboarding_dismissed';

interface OnboardingCardProps {
  onDismiss?: () => void;
}

export function OnboardingCard({ onDismiss }: OnboardingCardProps) {
  const navigate = useNavigate();

  const handleDismiss = () => {
    try {
      localStorage.setItem(ONBOARDING_DISMISSED_KEY, '1');
      onDismiss?.();
    } catch {
      // localStorage may be unavailable
    }
  };

  const guides = [
    {
      title: '您的项目',
      desc: '查看所有项目，点击卡片进入项目详情。',
      icon: <IconApps style={{ fontSize: 20, color: 'var(--zcid-color-primary)' }} />,
      action: '查看项目',
      onClick: () => navigate('/projects'),
    },
    {
      title: '创建第一个流水线',
      desc: '进入项目后，在流水线页面创建您的第一个 CI/CD 流水线。',
      icon: <IconPlayArrow style={{ fontSize: 20, color: 'var(--zcid-color-success)' }} />,
      action: '前往项目管理',
      onClick: () => navigate('/projects'),
    },
    {
      title: '文档',
      desc: '查阅使用说明和最佳实践。',
      icon: <IconLink style={{ fontSize: 20, color: 'var(--zcid-color-accent)' }} />,
      action: '查看文档',
      onClick: () => window.open('https://github.com/xjy/zcid', '_blank'),
    },
  ];

  return (
    <Card
      className="zcid-card"
      style={{ marginBottom: 'var(--zcid-space-section)', animation: 'fadeSlideUp var(--zcid-transition-slow) ease-out' }}
      title={
        <Typography.Text bold style={{ fontSize: 16 }}>
          欢迎使用 zcid
        </Typography.Text>
      }
    >
      <Space wrap size="large">
        {guides.map((g) => (
          <div
            key={g.title}
            className="zcid-card zcid-card-interactive"
            style={{ width: 240, padding: 'var(--zcid-space-card)' }}
            onClick={g.onClick}
          >
            <div style={{ marginBottom: 10 }}>{g.icon}</div>
            <Typography.Text bold style={{ display: 'block', marginBottom: 4 }}>{g.title}</Typography.Text>
            <Typography.Text
              type="secondary"
              style={{ fontSize: 13, lineHeight: '1.6', display: 'block', marginBottom: 10 }}
            >
              {g.desc}
            </Typography.Text>
            <Button type="text" size="small" style={{ padding: 0, color: 'var(--zcid-color-primary)' }}>
              {g.action}
            </Button>
          </div>
        ))}
      </Space>
      <div style={{ marginTop: 16, textAlign: 'right' }}>
        <Button
          size="small"
          type="text"
          onClick={handleDismiss}
          style={{ color: 'var(--zcid-color-text-tertiary)' }}
        >
          不再显示
        </Button>
      </div>
    </Card>
  );
}
