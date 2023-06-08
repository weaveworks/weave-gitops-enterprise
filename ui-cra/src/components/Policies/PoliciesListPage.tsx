import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { useListPolicies } from '../../contexts/PolicyViolations';
import { PolicyTable } from '@weaveworks/weave-gitops';

const Policies = () => {
  const { data, isLoading } = useListPolicies({});

  return (
    <PageTemplate documentTitle="Policies" path={[{ label: 'Policies' }]}>
      <ContentWrapper loading={isLoading} errors={data?.errors}>
        {data?.policies && <PolicyTable policies={data.policies} />}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default Policies;
