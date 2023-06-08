// import { useGetPolicyDetails } from '../../../contexts/PolicyViolations';
import { PolicyDetails, V2Routes } from '@weaveworks/weave-gitops';
import { useGetPolicyDetails } from '../../../contexts/PolicyViolations';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';

const PolicyDetailsPage = ({
  clusterName,
  id,
}: {
  clusterName: string;
  id: string;
}) => {
  const { data, isLoading, error } = useGetPolicyDetails({
    clusterName,
    policyName: id,
  });
  return (
    <PageTemplate
      documentTitle="Policies"
      path={[
        { label: 'Policies', url: V2Routes.Policies },
        { label: data?.policy?.name || '' },
      ]}
    >
      <ContentWrapper loading={isLoading}>
        <PolicyDetails policy={data?.policy || {}} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default PolicyDetailsPage;
