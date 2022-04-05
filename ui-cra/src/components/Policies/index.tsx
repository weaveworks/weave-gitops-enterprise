import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
import { PolicyTable } from './Table';
import { PolicyService } from './PolicyService';
import { useState } from 'react';
import { ListPoliciesResponse } from '../../capi-server/capi_server.pb';
import LoadingError from '../LoadingError';

const Policies = () => {
  const fetchPoliciesAPI = (payload: any = { page: 1, limit: 25 }) => {
    return PolicyService.getPolicyList().then((res: ListPoliciesResponse) => {
      res.total && setCount(res.total);
      return res;
    });
  };

  const [fetchPolicies] = useState(() => fetchPoliciesAPI);
  const [count, setCount] = useState<number>(0);
  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Policies">
        <SectionHeader
          className="count-header"
          path={[{ label: 'Policies', url: 'policies', count }]}
        />
        <ContentWrapper>
          <Title>Policies</Title>
          <LoadingError fetchFn={fetchPolicies}>
            {({ value }: { value: ListPoliciesResponse }) => (
              <>
                {value.total && value.total > 0 ? (
                  <PolicyTable policies={value.policies} />
                ) : (
                  <div>No data to display</div>
                )}
              </>
            )}
          </LoadingError>
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default Policies;
