import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
import { PolicyViolationsTable } from './Table';
import { useCallback, useContext, useState } from 'react';
import LoadingError from '../LoadingError';
import { EnterpriseClientContext } from '../../contexts/EnterpriseClient';
import {
  ListError,
  ListPolicyValidationsResponse,
  PolicyValidation,
} from '../../cluster-services/cluster_services.pb';

const PoliciesViolations = () => {
  const [count, setCount] = useState<number | undefined>(0);
  const { api } = useContext(EnterpriseClientContext);
  const [errors, SetErrors] = useState<ListError[] | undefined>();

  // const [payload, setPayload] = useState<any>({ page: 1, limit: 20, clusterId:'' });

  // Update payload on page change for next page request to work properly with pagination component in PolicyViolationTable component below
  // const updatePayload = (payload: any) => {
  //   setPayload(payload);
  // };
  const fetchPolicyViolationsAPI = useCallback(() => {
    return api.ListPolicyValidations({}).then(res => {
      !!res && setCount(res.total);
      !!res && SetErrors(res.errors);
      return res;
    });
    // TODO : Add pagination support for policy violations list API
    // Debendency: payload
  }, [api]);

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Violation Log">
        <SectionHeader
          className="count-header"
          path={[
            { label: 'Clusters', url: '/clusters', count },
            {
              label: 'Violation Log',
              url: 'violations',
              count: count,
            },
          ]}
        />
        <ContentWrapper errors={errors}>
          <Title>Violation Log</Title>
          <LoadingError fetchFn={fetchPolicyViolationsAPI}>
            {({ value }: { value: ListPolicyValidationsResponse }) => (
              <>
                {value.total && value.total > 0 ? (
                  <PolicyViolationsTable
                    violations={value.violations as PolicyValidation[]}
                  />
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
