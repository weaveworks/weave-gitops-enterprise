import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import CanariesList from './ListCanaries/CanariesList';
import OnboardingMessage from './Onboarding/OnboardingMessage';

import { Alert } from '@material-ui/lab';
import { LoadingPage } from '@weaveworks/weave-gitops';
import {
  useCanariesCount,
  useIsFlaggerAvailable,
} from '../../contexts/ProgressiveDelivery';
import { useApplicationsCount } from '../Applications/utils';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { SectionHeader } from '../Layout/SectionHeader';

const ProgressiveDelivery = () => {
  // To be discussed as we don't need to show counts for prev tabs in breadcrumb as it's not needed
  const applicationsCount = useApplicationsCount();
  const count = useCanariesCount();
  const {
    data: isFlaggerAvailable,
    isLoading,
    error,
  } = useIsFlaggerAvailable();

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

          {!isLoading && (
            <>{isFlaggerAvailable ? <CanariesList /> : <OnboardingMessage />}</>
          )}
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default ProgressiveDelivery;
