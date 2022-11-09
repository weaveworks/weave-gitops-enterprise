import { PageTemplate } from '../../Layout/PageTemplate';
import { ContentWrapper } from '../../Layout/ContentWrapper';

import HeaderSection from './HeaderSection';
import ParametersSection from './ParametersSection';
import { useGetPolicyDetails } from '../../../contexts/PolicyViolations';
import { Routes } from '../../../utils/nav';

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
    <PageTemplate
      documentTitle="Policies"
      path={[
        { label: 'Policies', url: Routes.Policies },
        { label: data?.policy?.name || '' },
      ]}
    >
      <ContentWrapper loading={isLoading}>
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
        />
        <ParametersSection parameters={policy?.parameters} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default PolicyDetails;
