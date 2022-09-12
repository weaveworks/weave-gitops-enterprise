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
    <PageTemplate documentTitle="WeGo · Delivery">
      {isLoading && <LoadingPage />}
      {error && <Alert severity="error">{error.message}</Alert>}

      {!isLoading && (
        <>{isFlaggerAvailable ? <CanariesList /> : <OnboardingMessage />}</>
      )}
    </PageTemplate>
  );
};

export default ProgressiveDelivery;
