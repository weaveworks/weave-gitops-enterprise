import { useGetPolicyConfigDetails } from '../../../contexts/PolicyConfigs';
import { Routes } from '../../../utils/nav';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
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
    <PageTemplate
      documentTitle="PolicyConfigs"
      path={[
        { label: 'PolicyConfigs', url: Routes.PolicyConfigs },
        { label: PolicyConfig?.name || '' },
      ]}
    >
      <ContentWrapper
        loading={isPolicyConfigLoading}
        warningMsg={
          PolicyConfig?.status === 'Warning'
            ? 'One or more than a policy isn’t found in the cluster.'
            : ''
        }
      >
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
      </ContentWrapper>
    </PageTemplate>
  );
};

export default PolicyConfigDetails;
