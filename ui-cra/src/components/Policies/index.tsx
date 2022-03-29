import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { CallbackStateContextProvider } from '@weaveworks/weave-gitops';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
import { PolicyTable } from './Table';
import { useState } from 'react';
import { Policy } from '../../types/custom';
import LoadingError from '../LoadingError';

const Policies = () => {
  const fetchUserApi = (payload: any = { page: 1, limit: 25 }) => {
    return Promise.resolve([
      {
        id: 'magalix.standards.soc2-type-i',
        name: 'SOC2 Type 1',
        category: 'Security',
        severity: 'high',
        createdAt: '2021-11-23T04:49:28.418Z',
      },
    ]).then(data => {
      setCount(data.length);
      return data;
    });
  };

  const [fetchUser, setFetchuser] = useState(() => fetchUserApi);
  const [count, setCount] = useState(0);
  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Policies">
        <CallbackStateContextProvider>
          <SectionHeader
            className="count-header"
            path={[{ label: 'Policies', url: 'policies', count }]}
          />
          <ContentWrapper>
            <Title>Policies</Title>
            <LoadingError fetchFn={fetchUser}>
              {({ value }: { value: Array<Policy> }) => (
                <>
                  <PolicyTable policies={value} />
                </>
              )}
            </LoadingError>
          </ContentWrapper>
        </CallbackStateContextProvider>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default Policies;
