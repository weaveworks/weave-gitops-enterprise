import CanariesList from './ListCanaries/CanariesList';
import OnboardingMessage from './Onboarding/OnboardingMessage';
import { Alert } from '@material-ui/lab';
import { useIsFlaggerAvailable } from '../../contexts/ProgressiveDelivery';
import { Routes } from '../../utils/nav';
import { Page } from '../Layout/App';

const ProgressiveDelivery = () => {
  const {
    data: isFlaggerAvailable,
    isLoading,
    error,
  } = useIsFlaggerAvailable();

  return (
    <>
      {!isLoading && isFlaggerAvailable ? (
        <Page loading={isLoading} path={[{ label: 'Progressive Delivery' }]}>
          {error && <Alert severity="error">{error?.message}</Alert>}
          <CanariesList />
        </Page>
      ) : (
        <Page loading={isLoading} path={[{ label: 'Progressive Delivery' }]}>
          {error && <Alert severity="error">{error.message}</Alert>}
          <OnboardingMessage />
        </Page>
      )}
    </>
  );
};

export default ProgressiveDelivery;
