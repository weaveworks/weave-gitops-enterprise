import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { useListPolicies } from '../../contexts/PolicyViolations';
import { PolicyTable } from '@weaveworks/weave-gitops';

const Policies = () => {
  const { data, isLoading } = useListPolicies({});

  return (
    <PageTemplate documentTitle="Policies" path={[{ label: 'Policies' }]}>
      <ContentWrapper loading={isLoading} errors={data?.errors}>
        {data?.policies && (
          <div id="policy-list">
            <PolicyTable policies={data.policies} />
          </div>
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default Policies;
