import * as React from 'react';
import { ContentWrapper } from './Layout/ContentWrapper';
import { PageTemplate } from './Layout/PageTemplate';
import { SectionHeader } from './Layout/SectionHeader';

export default class ErrorBoundary extends React.Component<
  any,
  { hasError: boolean; error: Error | null }
> {
  constructor(props: any) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error) {
    // Update state so the next render will show the fallback UI.
    return { hasError: true, error };
  }

  componentDidCatch(error: Error) {
    // You can also log the error to an error reporting service
    console.error(error);
  }

  render() {
    if (this.state.hasError) {
      // You can render any custom fallback UI
      return (
        <PageTemplate documentTitle="Error">
          <SectionHeader path={[{ label: 'Error' }]} />
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
