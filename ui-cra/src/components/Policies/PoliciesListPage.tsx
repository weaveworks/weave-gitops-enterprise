import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import { Page } from '../Layout/App';
import { useListPolicies } from '../../contexts/PolicyViolations';
import { PolicyTable } from '@weaveworks/weave-gitops';
import AuditAggregation from './AuditAggregation';

const Policies = () => {
  const { data, isLoading } = useListPolicies({});

  return (
    <Page loading={isLoading} path={[{ label: 'Policies' }]}>
      <NotificationsWrapper errors={data?.errors}>
        <AuditAggregation />
        {data?.policies && (
          <div id="policy-list">
            <PolicyTable policies={data.policies} />
          </div>
        )}
      </NotificationsWrapper>
    </Page>
  );
};

export default Policies;
