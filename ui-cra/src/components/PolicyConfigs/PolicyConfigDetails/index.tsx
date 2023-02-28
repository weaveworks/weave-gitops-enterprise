import { useGetPolicyConfigDetails } from '../../../contexts/PolicyConfigs';
import { Routes } from '../../../utils/nav';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import PolicyConfigHeaderSection from './PolicyConfigHeaderSection';

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
          { label: 'PolicyConfigs', url: Routes.Workspaces },
          { label: PolicyConfig?.name || '' },
        ]}
      >
        <ContentWrapper loading={isPolicyConfigLoading}>
          <PolicyConfigHeaderSection
            clusterName={PolicyConfig?.clusterName}
            age={PolicyConfig?.age}
          />
        </ContentWrapper>
      </PageTemplate>
    </>
  );
};

export default PolicyConfigDetails;
