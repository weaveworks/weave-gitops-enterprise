import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
import { CanaryTable } from './Table';
import { useCallback, useEffect, useState } from 'react';
import LoadingError from '../LoadingError';
import OnboardingMessage from './onboardingMessage';
import { Cached } from '@material-ui/icons';
import styled from 'styled-components';

const CounterWrapper = styled.div`
  display: flex;
  align-items: center;
  justify-content: start;
  position: absolute;
  color: #737373;
`;

const Blob = styled.div`
  .blob {
    background: red;
    border-radius: 50%;
    /* margin: 10px; */
    height: 8px;
    width: 8px;
    box-shadow: 0 0 0 0 #dd0a0a;
    transform: scale(1);
    animation: pulse 2s infinite;
    top: 0px;
    right: 0px;
  }

  @keyframes pulse {
    0% {
      transform: scale(0.95);
      box-shadow: 0 0 0 0 rgba(255, 0, 0, 0.7);
    }

    70% {
      transform: scale(1);
      box-shadow: 0 0 0 10px rgba(255, 0, 0, 0);
    }

    100% {
      transform: scale(0.95);
      box-shadow: 0 0 0 0 rgba(255, 0, 0, 0);
    }
  }
`;

const listCanaries = (): Promise<any> => {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        canaries: [
          {
            name: 'hello-world',
            clusterName: 'Default',
            namespace: 'default',
            provider: 'traefik',
            status: {
              phase: 'Succeeded',
              failedChecks: 1,
              canaryWeight: 0,
              iterations: 0,
              lastTransitionTime: '2022-05-11T13:54:51Z',
              conditions: [
                {
                  type: 'Promoted',
                  status: 'True',
                  lastUpdateTime: '2022-05-11T13:54:51Z',
                  lastTransitionTime: '2022-05-11T13:54:51Z',
                  reason: 'Succeeded',
                  message:
                    'Canary analysis completed successfully, promotion finished.',
                },
              ],
            },
            target_deployment: {
              uid: '',
              resource_version: '',
            },
            target_reference: {
              kind: 'Deployment',
              name: 'hello-world',
            },
          },
          {
            name: 'podinfo',
            namespace: 'podinfo',
            clusterName: 'Default',
            provider: 'traefik',
            status: {
              phase: 'Progressing',
              failedChecks: 1,
              canaryWeight: 15,
              iterations: 0,
              lastTransitionTime: '2022-05-11T13:54:51Z',
              conditions: [
                {
                  type: 'Promoted',
                  status: 'True',
                  lastUpdateTime: '2022-05-11T13:54:51Z',
                  lastTransitionTime: '2022-05-11T13:54:51Z',
                  reason: 'Progressing',
                  message:
                    'Canary analysis completed successfully, promotion finished.',
                },
              ],
            },
            target_deployment: {
              uid: '',
              resource_version: '',
            },
            target_reference: {
              kind: 'Deployment',
              name: 'hello-world',
            },
          },
          {
            name: 'backend',
            namespace: 'default',
            clusterName: 'Kind',
            provider: 'traefik',
            status: {
              phase: 'Failed',
              failedChecks: 0,
              canaryWeight: 0,
              iterations: 0,
              lastTransitionTime: '2022-05-11T13:54:51Z',
              conditions: [
                {
                  type: 'Promoted',
                  status: 'True',
                  lastUpdateTime: '2022-05-11T13:54:51Z',
                  lastTransitionTime: '2022-05-11T13:54:51Z',
                  reason: 'Failed',
                  message:
                    'Canary analysis completed successfully, promotion finished.',
                },
              ],
            },
            target_deployment: {
              uid: ' ',
              resource_version: '',
            },
            target_reference: {
              kind: 'Deployment',
              name: 'hello-world',
            },
          },
        ],
        total: 3,
        nextPageToken: 'looooong token',
        errors: [],
      });
    }, 1000);
  });
};

const ProgressiveDelivery = () => {
  const [count, setCount] = useState<number | undefined>(0);
  const [counter, setCounter] = useState(59);
  const [refetch, setRefetch] = useState<boolean>(false);

  const fetchCanariesAPI = useCallback(() => {
    if (refetch) {
      setCounter(59);
      setRefetch(false)
    };
    return listCanaries().then(res => {
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
          <Title>
            Canaries
            {/* <Blob>
              <div className="blob" />
            </Blob> */}
          </Title>

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
                  <OnboardingMessage />
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
