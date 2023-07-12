import { Flex } from '@weaveworks/weave-gitops';
import { useGetPolicyConfigDetails } from '../../../contexts/PolicyConfigs';
import { Routes } from '../../../utils/nav';
import { Page } from '../../Layout/App';
import { NotificationsWrapper } from '../../Layout/NotificationsWrapper';
import PolicyConfigHeaderSection from './PolicyConfigHeaderSection';
import PolicyDetailsCard from './PolicyDetailsCard';

const PolicyConfigDetails = ({
  clusterName,
  name,
}: {
  clusterName: string;
  name: string;
}) => {
  const { data: PolicyConfig, isLoading: isPolicyConfigLoading } =
    useGetPolicyConfigDetails({
      name,
      clusterName,
    });

  return (
    <Page
      loading={isPolicyConfigLoading}
      path={[
        { label: 'PolicyConfigs', url: Routes.PolicyConfigs },
        { label: PolicyConfig?.name || '' },
      ]}
    >
      <NotificationsWrapper
        warningMsg={
          PolicyConfig?.status === 'Warning'
            ? 'One or more than a policy isn’t found in the cluster.'
            : ''
        }
      >
        <Flex wide column gap="32">
          <PolicyConfigHeaderSection
            clusterName={PolicyConfig?.clusterName}
            age={PolicyConfig?.age}
            match={PolicyConfig?.match}
            matchType={PolicyConfig?.matchType}
          />
          <PolicyDetailsCard
            policies={PolicyConfig?.policies}
            totalPolicies={PolicyConfig?.totalPolicies}
            clusterName={clusterName}
          />
        </Flex>
      </NotificationsWrapper>
    </Page>
  );
};

export default PolicyConfigDetails;
