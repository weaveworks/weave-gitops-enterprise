import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import CanariesList from './ListCanaries/CanariesList';
import OnboardingMessage from './Onboarding/OnboardingMessage';

import { Alert } from '@material-ui/lab';
import { LoadingPage } from '@weaveworks/weave-gitops';
import { useIsFlaggerAvailable } from '../../contexts/ProgressiveDelivery';

const ProgressiveDelivery = () => {
  const {
    data: isFlaggerAvailable,
    isLoading,
    error,
  } = useIsFlaggerAvailable();

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Delivery">
        {isLoading && <LoadingPage />}
        {error && <Alert severity="error">{error.message}</Alert>}

        {!isLoading && (
          <>{isFlaggerAvailable ? <CanariesList /> : <OnboardingMessage />}</>
        )}
      </PageTemplate>
    </ThemeProvider>
  );
};

export default ProgressiveDelivery;
