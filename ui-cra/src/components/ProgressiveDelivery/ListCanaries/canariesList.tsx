import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { CanaryTable } from './Table';
import { useCallback, useEffect, useState } from 'react';
import LoadingError from '../../LoadingError';
import {
  ListCanariesResponse,
  ProgressiveDeliveryService,
} from '../../../cluster-services/prog.pb';
import { Canary } from '../../../cluster-services/types.pb';

const ProgressiveDelivery = () => {
  const [count, setCount] = useState<number | undefined>(0);
  const [counter, setCounter] = useState<number>(0);

  const fetchCanariesAPI = useCallback(() => {
    console.log(`counter call ${counter}`);
    return ProgressiveDeliveryService.ListCanaries({}).then(res => {
      !!res && setCount(res.canaries?.length || 0);
      return res;
    });
  }, [counter]);

  useEffect(() => {
    const intervalId = setInterval(() => {
      setCounter(prev => prev + 1);
    }, 60000);
    return () => {
      clearInterval(intervalId);
      setCounter(0);
    };
  }, []);
  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Canaries">
        <SectionHeader
          className="count-header"
          path={[
            { label: 'Applications', url: 'applications' },
            { label: 'Delivery', url: 'canaries', count },
          ]}
        />
        <ContentWrapper>
          <LoadingError fetchFn={fetchCanariesAPI}>
            {({ value }: { value: ListCanariesResponse }) => (
              <>
                {value.canaries?.length ? (
                  <CanaryTable canaries={value.canaries as Canary[]} />
                ) : (
                  <p>No data to display</p>
                )}
              </>
            )}
          </LoadingError>
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default ProgressiveDelivery;
