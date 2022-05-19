import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
import { PolicyViolationsTable } from './Table';
// import { PolicyService } from '../Policies/PolicyService';
import { useCallback, useContext, useState } from 'react';
import { ListPolicyValidationsResponse } from '../../capi-server/capi_server.pb';
import LoadingError from '../LoadingError';
import { EnterpriseClientContext } from '../../contexts/EnterpriseClient';

const PoliciesViolations = () => {
  const [count, setCount] = useState<number>(0);
  const { api } = useContext(EnterpriseClientContext);

  // const [payload, setPayload] = useState<any>({ page: 1, limit: 20, clusterId:'' });

  // Update payload on page change for next page request to work properly with pagination component in PolicyViolationTable component below
  // const updatePayload = (payload: any) => {
  //   setPayload(payload);
  // };
  // PolicyService.listPolicyViolations();
  const fetchPolicyViolationsAPI = useCallback(() => {
    return api
      .ListPolicyValidations({})
      .then((res: ListPolicyValidationsResponse | any) => {
        !!res && setCount(res.total);
        return res;
      });
    // TODO : Add pagination support for policy violations list API
    // Debendency: payload
  }, [api]);

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Violations Log">
        <SectionHeader
          className="count-header"
          path={[
            { label: 'Clusters', url: '/clusters' },
            { label: 'Violations Log', url: 'violations', count },
          ]}
        />
        <ContentWrapper>
          <Title>Violations Log</Title>
          <LoadingError fetchFn={fetchPolicyViolationsAPI}>
            {({ value }: { value: ListPolicyValidationsResponse }) => (
              <>
                {value.total && value.total > 0 ? (
                  <PolicyViolationsTable violations={value.violations} />
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

export default PoliciesViolations;
