import { useListPolicyValidations } from '../../contexts/PolicyViolations';
import LoadingWrapper from '../Workspaces/WorkspaceDetails/Tabs/WorkspaceTabsWrapper';
import { FieldsType, PolicyViolationsTable } from './Table';

const PoliciesViolations = ({ clusterName }: { clusterName: string }) => {
  const { data, error, isLoading } = useListPolicyValidations({ clusterName });
  return (
    <LoadingWrapper
      loading={isLoading}
      errorMessage={error?.message}
      errors={data?.errors}
    >
      <PolicyViolationsTable
        violations={data?.violations || []}
        tableType={FieldsType.policy}
      />
    </LoadingWrapper>
  );
};

export default PoliciesViolations;
