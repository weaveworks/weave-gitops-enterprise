import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
import { PolicyViolationsTable } from './Table';
import { PolicyService } from '../Policies/PolicyService';
import { useCallback, useState } from 'react';
import { ListPolicyValidationsResponse } from '../../cluster-services/cluster_services.pb';
import LoadingError from '../LoadingError';
import useClusters from '../../contexts/Clusters';

const PoliciesViolations = () => {
  const [policiesCount, setPoliciesCount] = useState<number>(0);
  const { count } = useClusters();

  // const [payload, setPayload] = useState<any>({ page: 1, limit: 20, clusterId:'' });

  // Update payload on page change for next page request to work properly with pagination component in PolicyViolationTable component below
  // const updatePayload = (payload: any) => {
  //   setPayload(payload);
  // };

  const fetchPolicyViolationsAPI = useCallback(() => {
    return PolicyService.listPolicyViolations().then(
      (res: ListPolicyValidationsResponse | any) => {
        !!res && setPoliciesCount(res.total);
        return res;
      },
    );
    // TODO : Add pagination support for policy violations list API
    // Debendency: payload
  }, []);

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Violations Log">
        <SectionHeader
          className="count-header"
          path={[
            { label: 'Clusters', url: '/clusters', count },
            {
              label: 'Violations Log',
              url: 'violations',
              count: policiesCount,
            },
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
