import CanariesList from './ListCanaries/CanariesList';
import OnboardingMessage from './Onboarding/OnboardingMessage';
import { Alert } from '@material-ui/lab';
import { Page } from '@weaveworks/weave-gitops';
import { useIsFlaggerAvailable } from '../../contexts/ProgressiveDelivery';
import { Routes } from '../../utils/nav';

const ProgressiveDelivery = () => {
  const {
    data: isFlaggerAvailable,
    isLoading,
    error,
  } = useIsFlaggerAvailable();

  return (
    <>
      {!isLoading && isFlaggerAvailable ? (
        <Page loading={isLoading} path={[]}>
          {error && <Alert severity="error">{error.message}</Alert>}
          <CanariesList />
        </Page>
      ) : (
        <Page
          loading={isLoading}
          path={[
            {
              label: 'Applications',
              url: Routes.Applications,
            },
            { label: 'Delivery' },
          ]}
        >
          {error && <Alert severity="error">{error.message}</Alert>}
          <OnboardingMessage />
        </Page>
      )}
    </>
  );
};

export default ProgressiveDelivery;
