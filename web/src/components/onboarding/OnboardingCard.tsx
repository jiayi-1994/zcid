import { Button } from '@arco-design/web-react';
import { IconApps, IconPlayArrow, IconBook } from '@arco-design/web-react/icon';
import { useNavigate } from 'react-router-dom';

export const ONBOARDING_DISMISSED_KEY = 'localStorage key for dismissing onboarding';

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
      title: '查看项目',
      desc: '浏览所有项目，点击卡片进入项目详情管理。',
      icon: <IconApps style={{ fontSize: 18 }} />,
      tone: 'stat-card-icon--primary',
      action: '前往项目',
      onClick: () => navigate('/projects'),
    },
    {
      title: '创建流水线',
      desc: '进入项目后，创建您的第一个 CI/CD 流水线。',
      icon: <IconPlayArrow style={{ fontSize: 18 }} />,
      tone: 'stat-card-icon--success',
      action: '开始创建',
      onClick: () => navigate('/projects'),
    },
    {
      title: '查看文档',
      desc: '查阅使用说明和最佳实践，快速上手。',
      icon: <IconBook style={{ fontSize: 18 }} />,
      tone: 'stat-card-icon--warning',
      action: '阅读文档',
      onClick: () => window.open('https://github.com/xjy/zcid', '_blank'),
    },
  ];

  return (
    <div className="onboarding-card">
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          marginBottom: 4,
        }}
      >
        <h3 className="onboarding-title">欢迎使用 zcid</h3>
        <Button size="small" type="text" onClick={handleDismiss}>
          不再显示
        </Button>
      </div>
      <p className="onboarding-desc">你的 CI/CD 旅程从这里开始，跟随引导快速上手</p>
      <div className="onboarding-steps">
        {guides.map((g) => (
          <div key={g.title} className="onboarding-step" onClick={g.onClick}>
            <div className={`stat-card-icon ${g.tone}`} style={{ marginBottom: 8 }}>
              {g.icon}
            </div>
            <div className="onboarding-step-title">{g.title}</div>
            <div className="onboarding-step-desc">{g.desc}</div>
            <Button
              type="text"
              size="mini"
              style={{ padding: 0, marginTop: 8, color: 'var(--primary)', fontWeight: 500 }}
            >
              {g.action} &rarr;
            </Button>
          </div>
        ))}
      </div>
    </div>
  );
}
