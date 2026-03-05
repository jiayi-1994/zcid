import { Button } from '@arco-design/web-react';
import { type ErrorInfo, type ReactNode, Component } from 'react';

interface ErrorBoundaryProps {
  children: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = {
    hasError: false,
  };

  static getDerivedStateFromError(): ErrorBoundaryState {
    return { hasError: true };
  }

  componentDidCatch(_error: Error, _errorInfo: ErrorInfo) {
    // no-op: logging will be wired in later stories
  }

  handleRetry = () => {
    this.setState({ hasError: false });
  };

  render() {
    if (this.state.hasError) {
      return (
        <div
          role="alert"
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            minHeight: '60vh',
            padding: 48,
            textAlign: 'center',
            animation: 'fadeIn var(--zcid-transition-slow) ease-out',
          }}
        >
          <div
            style={{
              width: 64,
              height: 64,
              borderRadius: 'var(--zcid-radius-full)',
              background: 'var(--zcid-color-bg-tertiary)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              marginBottom: 20,
              fontSize: 28,
              color: 'var(--zcid-color-text-tertiary)',
            }}
          >
            !
          </div>
          <h2 style={{ margin: '0 0 8px', color: 'var(--zcid-color-text-primary)', fontWeight: 600 }}>
            页面出现错误
          </h2>
          <p style={{ margin: '0 0 24px', color: 'var(--zcid-color-text-secondary)', fontSize: 14 }}>
            请重试或稍后再试。
          </p>
          <Button type="primary" onClick={this.handleRetry}>
            重试
          </Button>
        </div>
      );
    }

    return this.props.children;
  }
}
