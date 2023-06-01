import { NotificationsWrapper } from '../../Layout/NotificationsWrapper';
import HeaderSection from './HeaderSection';
import ParametersSection from './ParametersSection';
import { useGetPolicyDetails } from '../../../contexts/PolicyViolations';
import { Routes } from '../../../utils/nav';
import { Page } from '@weaveworks/weave-gitops';

const PolicyDetails = ({
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
  const policy = data?.policy;

  return (
    <Page
      loading={isLoading}
      path={[
        { label: 'Policies', url: Routes.Policies },
        { label: data?.policy?.name || '' },
      ]}
    >
      <NotificationsWrapper>
        <HeaderSection
          id={policy?.id}
          clusterName={policy?.clusterName}
          tags={policy?.tags}
          severity={policy?.severity}
          category={policy?.category}
          targets={policy?.targets}
          description={policy?.description}
          howToSolve={policy?.howToSolve}
          code={policy?.code}
          tenant={policy?.tenant}
          modes={policy?.modes}
        />
        <ParametersSection parameters={policy?.parameters} />
      </NotificationsWrapper>
    </Page>
  );
};

export default PolicyDetails;
