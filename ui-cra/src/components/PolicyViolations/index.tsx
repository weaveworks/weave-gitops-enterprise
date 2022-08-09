import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { FieldsType, PolicyViolationsTable } from './Table';
import useClusters from './../../contexts/Clusters';
import { useListPolicyValidations } from '../../contexts/PolicyViolations';
import { Alert } from '@material-ui/lab';
import { LoadingPage } from '@weaveworks/weave-gitops';

const PoliciesViolations = () => {
  const clustersCount = useClusters().count;
  const { data, isLoading, error } = useListPolicyValidations({});

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Violation Log">
        <SectionHeader
          className="count-header"
          path={[
            { label: 'Clusters', url: '/clusters', count: clustersCount },
            {
              label: 'Violation Log',
              url: 'violations',
              count: data?.total,
            },
          ]}
        />
        <ContentWrapper errors={data?.errors}>
          {isLoading && <LoadingPage />}
          {error && <Alert severity="error">{error.message}</Alert>}
          {data?.violations && (
            <PolicyViolationsTable
              violations={data?.violations || []}
              tableType={FieldsType.policy}
            />
          )}
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default PoliciesViolations;
