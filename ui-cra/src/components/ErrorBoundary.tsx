import React, { useState, useEffect, FC } from 'react';
import { useLocation } from 'react-router-dom';
import { ContentWrapper } from './Layout/ContentWrapper';
import { PageTemplate } from './Layout/PageTemplate';

interface Props {
  hasError: boolean;
  error: Error | null;
}

class ErrorBoundaryDetail extends React.Component<any, Props> {
  constructor(props: any) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  componentDidUpdate(prevProps: Props) {
    if (!this.props.hasError && prevProps.hasError) {
      this.setState({ hasError: false });
    }
  }

  componentDidCatch(error: Error) {
    console.error(error);
    this.props.setHasError(true);
  }

  render() {
    if (this.state.hasError) {
      return (
        <PageTemplate
          documentTitle="Error"
          path={[
            {
              label: 'Error',
            },
          ]}
        >
          <ContentWrapper>
            <h3>Something went wrong.</h3>
            <pre>{this.state.error?.message}</pre>
            <pre>{this.state.error?.stack}</pre>
          </ContentWrapper>
        </PageTemplate>
      );
    }

    return this.props.children;
  }
}

/** Function component wrapper as we need useEffect to set the state back to false on location changing **/
const ErrorBoundary: FC<{
  children: React.ReactNode;
}> = ({ children }) => {
  const [hasError, setHasError] = useState<boolean>(false);
  const location = useLocation();

  useEffect(() => {
    if (hasError) {
      setHasError(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [location.key]);

  return (
    <ErrorBoundaryDetail hasError={hasError} setHasError={setHasError}>
      {children}
    </ErrorBoundaryDetail>
  );
};

export default ErrorBoundary;
