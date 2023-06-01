import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import { PolicyTable } from './Table';
import { useListListPolicies } from '../../contexts/PolicyViolations';
import { Page } from '@weaveworks/weave-gitops';

const Policies = () => {
  const { data, isLoading } = useListListPolicies({});

  return (
    <Page loading={isLoading} path={[{ label: 'Policies' }]}>
      <NotificationsWrapper errors={data?.errors}>
        {data?.policies && <PolicyTable policies={data.policies} />}
      </NotificationsWrapper>
    </Page>
  );
};

export default Policies;
