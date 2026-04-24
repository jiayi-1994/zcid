import { type ErrorInfo, type ReactNode, Component } from 'react';
import { Btn } from '../ui/Btn';

interface ErrorBoundaryProps {
  children: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = { hasError: false };

  static getDerivedStateFromError(): ErrorBoundaryState {
    return { hasError: true };
  }

  componentDidCatch(_error: Error, _errorInfo: ErrorInfo) {
    // logging wired later
  }

  handleRetry = () => { this.setState({ hasError: false }); };

  render() {
    if (this.state.hasError) {
      return (
        <div className="zc" style={{ minHeight: '60vh' }}>
          <div
            role="alert"
            style={{
              display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
              minHeight: '60vh', padding: 48, textAlign: 'center',
              fontFamily: 'var(--font-sans)',
            }}
          >
            <div style={{
              width: 56, height: 56, borderRadius: '50%',
              background: 'var(--red-soft)', color: 'var(--red-ink)',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              marginBottom: 18, fontSize: 26, fontWeight: 600,
            }}>
              !
            </div>
            <h2 style={{ margin: '0 0 6px', color: 'var(--z-900)', fontWeight: 600, fontSize: 18 }}>
              页面出现错误
            </h2>
            <p style={{ margin: '0 0 20px', color: 'var(--z-500)', fontSize: 13 }}>
              请重试或稍后再试。
            </p>
            <Btn variant="primary" onClick={this.handleRetry}>重试</Btn>
          </div>
        </div>
      );
    }
    return this.props.children;
  }
}
