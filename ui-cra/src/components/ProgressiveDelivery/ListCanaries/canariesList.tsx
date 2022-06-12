import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { ContentWrapper, Title } from '../../Layout/ContentWrapper';
import { CanaryTable } from './Table';
import { useCallback, useEffect, useState } from 'react';
import LoadingError from '../../LoadingError';
import { Cached } from '@material-ui/icons';
import styled from 'styled-components';
import { CanaryService } from '../CanaryService';

const CounterWrapper = styled.div`
  display: flex;
  align-items: center;
  justify-content: start;
  position: absolute;
  color: #737373;
`;

const ProgressiveDelivery = () => {
  const [count, setCount] = useState<number | undefined>(0);
  const [counter, setCounter] = useState(59);
  const [refetch, setRefetch] = useState<boolean>(false);

  const fetchCanariesAPI = useCallback(() => {
    if (refetch) {
      setCounter(59);
      setRefetch(false);
    }
    return CanaryService.listCanaries().then(res => {
      !!res && setCount(res.total);
      return res;
    });
  }, [refetch]);

  useEffect(() => {
    const intervalId = setInterval(() => {
      setCounter(counter - 1);
      if (counter === 1) {
        setCounter(59);
        setRefetch(true);
      }
    }, 1000);
    return () => clearInterval(intervalId);
  }, [counter]);
  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Canaries">
        <SectionHeader
          className="count-header"
          path={[
            { label: 'Applications', url: 'applications' },
            { label: 'Canaries', url: 'canaries', count },
          ]}
        />
        <ContentWrapper>
          <Title>Canaries</Title>

          <LoadingError fetchFn={fetchCanariesAPI}>
            {({ value }: { value: any }) => (
              <>
                {value.total && value.total > 0 ? (
                  <>
                    <CounterWrapper>
                      <p>Updating in {counter} seconds...</p>
                      <Cached onClick={() => setRefetch(true)} />
                    </CounterWrapper>
                    <CanaryTable canaries={value.canaries as any[]} />
                  </>
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
