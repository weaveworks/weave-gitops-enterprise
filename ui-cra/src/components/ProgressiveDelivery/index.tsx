import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import { useCallback, useState } from 'react';
import CanariesList from './ListCanaries/CanariesList';
import OnboardingMessage from './Onboarding/OnboardingMessage';

import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { useQuery } from 'react-query';
import { LoadingPage } from '@weaveworks/weave-gitops';
import { Alert } from '@material-ui/lab';
import {
  IsFlaggerAvailableResponse,
  ProgressiveDeliveryService,
} from '@weaveworks/progressive-delivery';

interface FlaggerStatus {
  isFlaggerAvailabl: boolean;
}

const ProgressiveDelivery = () => {
  const [count, setCount] = useState<number | undefined>();
  const { error, data, isLoading } = useQuery<FlaggerStatus, Error>(
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
            { label: 'Applications', url: '/applications' },
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