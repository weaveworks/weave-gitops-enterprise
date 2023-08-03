import {
    Flex,
    LoadingPage,
    PolicyTable
} from '@weaveworks/weave-gitops';
import { useListPolicies } from '../../contexts/PolicyViolations';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';

export const PoliciesTab = () => {
  const { data, isLoading } = useListPolicies({});

  return isLoading ? (
    <LoadingPage />
  ) : (
    <NotificationsWrapper errors={data?.errors}>
      {data?.policies && (
        <Flex wide id="policy-list">
          <PolicyTable policies={data.policies} />
        </Flex>
      )}
    </NotificationsWrapper>
  );
};
