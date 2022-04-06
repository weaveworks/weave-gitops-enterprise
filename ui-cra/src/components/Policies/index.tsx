import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
import { PolicyTable } from './Table';
import { PolicyService } from './PolicyService';
import { useCallback } from 'react';
import { ListPoliciesResponse } from '../../capi-server/capi_server.pb';
import LoadingError, { useRequest } from '../LoadingError';

const Policies = () => {
  const fetchPolicies = useCallback(
    (payload: any = { page: 1, limit: 25 }) => PolicyService.getPolicyList(),
    [],
  );

  const requestInfo = useRequest(fetchPolicies);
  const data = requestInfo.data as ListPoliciesResponse;
  const count = data ? data.total : 0;
  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Policies">
        <SectionHeader
          className="count-header"
          path={[{ label: 'Policies', url: 'policies', count }]}
        />
        <ContentWrapper>
          <Title >Policies</Title>
          <LoadingError requestInfo={requestInfo}>
            {data?.total! > 0 ? (
              <PolicyTable policies={data.policies} />
            ) : (
              <div>No data to display</div>
            )}
          </LoadingError>
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default Policies;
