import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { CallbackStateContextProvider } from '@weaveworks/weave-gitops';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
import { PolicyTable } from './Table';
import { PolicyService } from './PolicyService';
import { useState } from 'react';
import { Policy } from '../../types/custom';
import LoadingError from '../LoadingError';

interface IPolicyResponse {
  policies: Array<Policy>;
  total: number;
}

const Policies = () => {
  const fetchPoliciesAPI = (payload: any = { page: 1, limit: 25 }) => {
    return PolicyService.getPolicyList().then((res: IPolicyResponse) => {
      setCount(res.total);
      return res;
    });
  };

  const [fetchPolicies] = useState(() => fetchPoliciesAPI);
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
            <LoadingError fetchFn={fetchPolicies}>
              {({ value }: { value: IPolicyResponse }) => (
                <>
                  {value.total > 0 ? (
                    <PolicyTable policies={value.policies} />
                  ) : (
                    <div>No data to display</div>
                  )}
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
