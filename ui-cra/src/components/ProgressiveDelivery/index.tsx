import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import { useCallback, useState } from 'react';
import LoadingError from '../LoadingError';
import CanariesList from './ListCanaries/canariesList';
import OnboardingMessage from './Onboarding/onboardingMessage';
import {
  IsFlaggerAvailableResponse,
  ProgressiveDeliveryService,
} from '../../cluster-services/prog.pb';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';

const ProgressiveDelivery = () => {
  const [count, setCount] = useState<number | undefined>();

  const isFlaggerInstalledAPI = useCallback(() => {
    return ProgressiveDeliveryService.IsFlaggerAvailable({}).then(
      ({ clusters }: IsFlaggerAvailableResponse) => {
        if (clusters === undefined || Object.keys(clusters).length === 0)
          return false;
        else {
          return Object.values(clusters).some(
            (value: boolean) => value === true,
          );
        }
      },
    );
  }, []);
  const onCountChange = useCallback((count: number) => {
    console.log(count);
    setCount(count);
  }, []);

  return (
    <div
      style={{
        height: '100vh',
        display: 'flex',
      }}
    >
      <ThemeProvider theme={localEEMuiTheme}>
        <PageTemplate documentTitle="WeGo · Delivery">
          <SectionHeader
            className="count-header"
            path={[
              { label: 'Applications', url: 'applications' },
              { label: 'Delivery', url: 'canaries', count },
            ]}
          />
          <ContentWrapper>
            <LoadingError fetchFn={isFlaggerInstalledAPI}>
              {({ value }: { value: boolean }) => (
                <>
                  {value ? (
                    <CanariesList onCountChange={onCountChange} />
                  ) : (
                    <OnboardingMessage />
                  )}
                </>
              )}
            </LoadingError>
          </ContentWrapper>
        </PageTemplate>
      </ThemeProvider>
    </div>
  );
};

export default ProgressiveDelivery;
