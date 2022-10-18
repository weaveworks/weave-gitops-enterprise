import { PageTemplate } from '../Layout/PageTemplate';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { FieldsType, PolicyViolationsTable } from './Table';
import { useListPolicyValidations } from '../../contexts/PolicyViolations';
import { Routes } from '../../utils/nav';

const PoliciesViolations = () => {
  const { data, isLoading, error } = useListPolicyValidations({});
  return (
    <PageTemplate
      documentTitle="Violation Log"
      path={[
        { label: 'Clusters', url: Routes.Clusters },
        {
          label: 'Violation Log',
          count: data?.total,
        },
      ]}
    >
      <ContentWrapper
        loading={isLoading}
        errorMessage={error?.message}
        errors={data?.errors}
      >
        <PolicyViolationsTable
          violations={data?.violations || []}
          tableType={FieldsType.policy}
        />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default PoliciesViolations;
