import { useGetPolicyConfigDetails } from '../../../contexts/PolicyConfigs';
import { Routes } from '../../../utils/nav';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import { WarningIcon, WarningWrapper } from '../PolicyConfigStyles';
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
    <>
      <PageTemplate
        documentTitle="PolicyConfigs"
        path={[
          { label: 'PolicyConfigs', url: Routes.PolicyConfigs },
          { label: PolicyConfig?.name || '' },
        ]}
      >
        {PolicyConfig?.status === 'Warning' && (
          <WarningWrapper
            severity="warning"
            iconMapping={{
              warning: <WarningIcon />,
            }}
          >
            <span>One or more than a policy isn’t found in the cluster</span>
          </WarningWrapper>
        )}

        <ContentWrapper
          loading={isPolicyConfigLoading}
          customMaxHieght={PolicyConfig?.status === 'Warning' ? "calc(100vh - 142px)" : undefined}
        >
          <PolicyConfigHeaderSection
            clusterName={PolicyConfig?.clusterName}
            age={PolicyConfig?.age}
            match={PolicyConfig?.match}
          />
          <PolicyDetailsCard
            policies={PolicyConfig?.policies}
            totalPolicies={PolicyConfig?.totalPolicies}
            clusterName={clusterName}
          />
        </ContentWrapper>
      </PageTemplate>
    </>
  );
};

export default PolicyConfigDetails;
