import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import { useCallback, useState } from 'react';
import CanariesList from './ListCanaries/CanariesList';
import OnboardingMessage from './Onboarding/OnboardingMessage';

import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { LoadingPage } from '@weaveworks/weave-gitops';
import { Alert } from '@material-ui/lab';

import { useApplicationsCount } from '../Applications/utils';
import { useIsFlaggerAvailable } from '../../hooks/progressiveDelivery';



const ProgressiveDelivery = () => {
  // To be discussed as we don't need to show counts for prev tabs in breadcrumb as it's not needed
  const applicationsCount = useApplicationsCount();
  
  const [count, setCount] = useState<number | undefined>();
  const { error, data, isLoading } = useIsFlaggerAvailable();

  const onCountChange = useCallback((count: number) => {
    setCount(count);
  }, []);

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Delivery">
        <SectionHeader
          className="count-header"
          path={[
            {
              label: 'Applications',
              url: '/applications',
              count: applicationsCount,
            },
            { label: 'Delivery', count },
          ]}
        />
        <ContentWrapper>
          {isLoading && <LoadingPage />}
          {error && <Alert severity="error">{error.message}</Alert>}

          {!!data && (
            <>
              {data.isFlaggerAvailable ? (
                <CanariesList onCountChange={onCountChange} />
              ) : (
                <OnboardingMessage />
              )}
            </>
          )}
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default ProgressiveDelivery;
