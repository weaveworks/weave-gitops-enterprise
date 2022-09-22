import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PolicyTable } from './Table';
import { useListListPolicies } from '../../contexts/PolicyViolations';
import { Alert } from '@material-ui/lab';
import { LoadingPage } from '@weaveworks/weave-gitops';

const Policies = () => {
  const { data, isLoading, error } = useListListPolicies({});

  return (
      <PageTemplate documentTitle="WeGO · Policies">
        <SectionHeader
          className="count-header"
          path={[{ label: 'Policies', url: 'policies', count: data?.total }]}
        />
        <ContentWrapper errors={data?.errors}>
          {isLoading && <LoadingPage />}
          {error && <Alert severity="error">{error.message}</Alert>}
          {data?.policies && <PolicyTable policies={data.policies} />}
        </ContentWrapper>
      </PageTemplate>
  );
};

export default Policies;
