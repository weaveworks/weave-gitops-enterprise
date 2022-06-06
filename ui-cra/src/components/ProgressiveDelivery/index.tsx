import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import { useCallback } from 'react';
import LoadingError from '../LoadingError';
import OnboardingMessage from './onboardingMessage';
import CanariesList from './canariesList';
const getFlaggerStatus = (): Promise<any> => {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        Default: true,
        'LeafCluster-1': false,
        'LeafCluster-2': false,
        'LeafCluster-3': false,
      });
    }, 1000);
  });
};

const ProgressiveDelivery = () => {
  const isFlaggerInstalledAPI = useCallback(() => {
    return getFlaggerStatus().then((res: Object) => {
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
