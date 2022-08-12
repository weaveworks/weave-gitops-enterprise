import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { ContentWrapper, Title } from '../../Layout/ContentWrapper';
import { Alert } from '@material-ui/lab';
import { LoadingPage } from '@weaveworks/weave-gitops';
import ViolationDetails from './ViolationDetails';
import useClusters from '../../../contexts/Clusters';
import {
  useCountPolicyValidations,
  useGetPolicyValidationDetails,
} from '../../../contexts/PolicyViolations';

const PolicyViolationDetails = ({
  id,
  clusterName,
}: {
  id: string;
  clusterName: string;
}) => {
  const { count } = useClusters();
  const policyViolationsCount = useCountPolicyValidations({});
  const { data, error, isLoading } = useGetPolicyValidationDetails({
    clusterName,
    violationId: id,
  });
  const name = data?.violation?.name || '';
  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Violation Logs">
        <SectionHeader
          className="count-header"
          path={[
            { label: 'Clusters', url: '/clusters', count },
            {
              label: 'Violation Logs',
              url: '/clusters/violations',
              count: policyViolationsCount,
            },
            { label: name },
          ]}
        />
        <ContentWrapper>
          <Title>{name}</Title>
          {isLoading && <LoadingPage />}
          {error && <Alert severity="error">{error.message}</Alert>}
          {data?.violation && <ViolationDetails violation={data.violation} />}
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default PolicyViolationDetails;
