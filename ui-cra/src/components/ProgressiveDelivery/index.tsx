import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import { useCallback, useState } from 'react';
import CanariesList from './ListCanaries/canariesList';
import OnboardingMessage from './Onboarding/OnboardingMessage';
import {
  IsFlaggerAvailableResponse,
  ProgressiveDeliveryService,
} from '../../cluster-services/prog.pb';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { useQuery } from 'react-query';
import { LoadingPage } from '@weaveworks/weave-gitops';
import { Alert } from '@material-ui/lab';

interface IFlaggerStatus {
  isFlaggerAvailabl: boolean;
}

const ProgressiveDelivery = () => {
  const [count, setCount] = useState<number | undefined>();
  const { error, data, isLoading } = useQuery<IFlaggerStatus, Error>(
    'flagger',
    () =>
      ProgressiveDeliveryService.IsFlaggerAvailable({}).then(
        ({ clusters }: IsFlaggerAvailableResponse) => {
          if (clusters === undefined || Object.keys(clusters).length === 0)
            return { isFlaggerAvailabl: false };
          else {
            return {
              isFlaggerAvailabl: Object.values(clusters).some(
                (value: boolean) => value === true,
              ),
            };
          }
        },
      ),
  );
  const onCountChange = useCallback((count: number) => {
    setCount(count);
  }, []);

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Delivery">
        <SectionHeader
          className="count-header"
          path={[
            { label: 'Applications', url: 'applications' },
            { label: 'Delivery', count },
          ]}
        />
        <ContentWrapper>
          {isLoading && <LoadingPage />}
          {error && <Alert severity="error">{error.message}</Alert>}

          {!!data && (
            <>
              {data.isFlaggerAvailabl ? (
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
