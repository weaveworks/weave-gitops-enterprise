import { PolicyTable } from './Table';
import { useListListPolicies } from '../../contexts/PolicyViolations';
import { Page } from '@weaveworks/weave-gitops';

const Policies = () => {
  const { data, isLoading } = useListListPolicies({});

  return (
    <Page
      loading={isLoading}
      error={data?.errors}
      path={[{ label: 'Policies' }]}
    >
      {data?.policies && <PolicyTable policies={data.policies} />}
    </Page>
  );
};

export default Policies;
