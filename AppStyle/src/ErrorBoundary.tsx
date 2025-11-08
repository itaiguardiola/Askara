import { Component, ReactNode, ErrorInfo } from 'react';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null, errorInfo: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error, errorInfo: null };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
    this.setState({ error, errorInfo });
  }

  render() {
    if (this.state.hasError) {
      return (
        <div style={{
          padding: '20px',
          fontFamily: 'monospace',
          backgroundColor: '#fee',
          color: '#c00',
          minHeight: '100vh'
        }}>
          <h1>Something went wrong</h1>
          <details style={{ whiteSpace: 'pre-wrap', marginTop: '20px' }}>
            <summary>Error Details</summary>
            <p><strong>Error:</strong> {this.state.error?.toString()}</p>
            <p><strong>Stack:</strong></p>
            <pre>{this.state.error?.stack}</pre>
            {this.state.errorInfo && (
              <>
                <p><strong>Component Stack:</strong></p>
                <pre>{this.state.errorInfo.componentStack}</pre>
              </>
            )}
          </details>
        </div>
      );
    }

    return this.props.children;
  }
}
