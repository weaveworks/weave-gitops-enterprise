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
import { Breadcrumb } from '../../Breadcrumbs';

const PolicyViolationDetails = ({
  id,
  clusterName,
  source,
  sourcePath,
}: {
  id: string;
  clusterName: string;
  source?: string;
  sourcePath?: string;
}) => {
  const { count } = useClusters();
  const policyViolationsCount = useCountPolicyValidations({});
  const { data, error, isLoading } = useGetPolicyValidationDetails({
    clusterName,
    violationId: id,
  });
  const { message, namespace, entity } = data?.violation || {
    message: '',
    namespace: '',
    entity: '',
  };

  const headerPath: Breadcrumb[] = !!source
    ? [
        { label: 'Applications', url: '/applications', count },
        {
          label: entity || '',
          url: `/${sourcePath}/violations?clusterName=${clusterName}&name=${entity}&namespace=${namespace}`,
          count: policyViolationsCount,
        },
        { label: message || '' },
      ]
    : [
        { label: 'Clusters', url: '/clusters', count },
        {
          label: 'Violation Logs',
          url: '/clusters/violations',
          count: policyViolationsCount,
        },
        { label: message || '' },
      ];
  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Violation Logs">
        <SectionHeader className="count-header" path={headerPath} />
        <ContentWrapper>
          <Title>{message}</Title>
          {isLoading && <LoadingPage />}
          {error && <Alert severity="error">{error.message}</Alert>}
          {data?.violation && (
            <ViolationDetails violation={data.violation} source={source} />
          )}
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default PolicyViolationDetails;
