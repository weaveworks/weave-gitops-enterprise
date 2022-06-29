import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { ContentWrapper, Title } from '../../Layout/ContentWrapper';
import { useCallback, useState } from 'react';
import LoadingError from '../../LoadingError';
import { useParams } from 'react-router-dom';
import { GetPolicyValidationResponse } from '../../../cluster-services/cluster_services.pb';
import { PolicyService } from '../../Policies/PolicyService';
import ViolationDetails from './ViolationDetails';
import useClusters from '../../../contexts/Clusters';

const PolicyViolationDetails = () => {
  const { id } = useParams<{ id: string }>();
  const { count } = useClusters();
  const [name, setName] = useState('');
  const { clusterName } = useParams<{ clusterName: string }>();

  const fetchPolicyViolationById = useCallback(
    () =>
      PolicyService.getPolicyViolationById(id, clusterName).then(
        (res: GetPolicyValidationResponse) => {
          res.violation?.name && setName(res.violation.name);
          return res;
        },
      ),
    [id, clusterName],
  );

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Violation Logs">
        <SectionHeader
          className="count-header"
          path={[
            { label: 'Clusters', url: '/clusters', count },
            { label: 'Violation Logs', url: '/clusters/violations' },
            { label: name, url: 'policy-violation-details' },
          ]}
        />
        <ContentWrapper>
          <Title>{name}</Title>
          <LoadingError fetchFn={fetchPolicyViolationById}>
            {({
              value: { violation },
            }: {
              value: GetPolicyValidationResponse;
            }) => (
              <>
                <ViolationDetails violation={violation}></ViolationDetails>
              </>
            )}
          </LoadingError>
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default PolicyViolationDetails;
