import { PageTemplate } from '../Layout/PageTemplate';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PolicyTable } from './Table';
import { useListListPolicies } from '../../contexts/PolicyViolations';
import { Routes } from '../../utils/nav';

const Policies = () => {
  const { data, isLoading, error } = useListListPolicies({});

  return (
    <PageTemplate
      documentTitle="Policies"
      path={[{ label: 'Policies', url: Routes.Policies }]}
    >
      <ContentWrapper
        loading={isLoading}
        notifications={[
          {
            message: { text: error?.message },
            severity: 'error',
          },
        ]}
        errors={data?.errors}
      >
        {data?.policies && <PolicyTable policies={data.policies} />}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default Policies;
