import { PageTemplate } from '../Layout/PageTemplate';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PolicyTable } from './Table';
import { useListListPolicies } from '../../contexts/PolicyViolations';

const Policies = () => {
  const { data, isLoading, error } = useListListPolicies({});

  return (
    <PageTemplate documentTitle="Policies" path={[{ label: 'Policies', url: 'policies', count: data?.total }]}>
      <ContentWrapper
        loading={isLoading}
        errorMessage={error?.message}
        errors={data?.errors}
      >
        {data?.policies && <PolicyTable policies={data.policies} />}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default Policies;
