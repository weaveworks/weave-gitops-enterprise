import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import { useCallback } from 'react';
import LoadingError from '../LoadingError';
import CanariesList from './ListCanaries/canariesList';
import OnboardingMessage from './Onboarding/onboardingMessage';
import { CanaryService } from './CanaryService';

const ProgressiveDelivery = () => {
  const isFlaggerInstalledAPI = useCallback(() => {
    return CanaryService.getFlaggerStatus().then((res: Object) => {
      if (Object.keys(res).length === 0) return false;
      return Object.values(res).some((value: boolean) => value === true);
    });
  }, []);

  return (
    <div
      style={{
        height: '100vh',
        display: 'flex',
      }}
    >
      <ThemeProvider theme={localEEMuiTheme}>
        <PageTemplate documentTitle="WeGo · Canaries">
          <LoadingError fetchFn={isFlaggerInstalledAPI}>
            {({ value }: { value: boolean }) => (
              <>{value ? <CanariesList /> : <OnboardingMessage />}</>
            )}
          </LoadingError>
        </PageTemplate>
      </ThemeProvider>
    </div>
  );
};

export default ProgressiveDelivery;
