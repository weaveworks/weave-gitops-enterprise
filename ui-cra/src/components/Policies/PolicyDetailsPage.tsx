import { Page, PolicyDetails, V2Routes } from '@weaveworks/weave-gitops';
import { useGetPolicyDetails } from '../../contexts/PolicyViolations';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';

const PolicyDetailsPage = ({
  clusterName,
  id,
}: {
  clusterName: string;
  id: string;
}) => {
  const { data, isLoading } = useGetPolicyDetails({
    clusterName,
    policyName: id,
  });

  return (
    <Page
      loading={isLoading}
      path={[
        { label: 'Policies', url: V2Routes.Policies },
        { label: data?.policy?.name || '' },
      ]}
    >
      <NotificationsWrapper>
        <PolicyDetails policy={data?.policy || {}} />
      </NotificationsWrapper>
    </Page>
  );
};

export default PolicyDetailsPage;
