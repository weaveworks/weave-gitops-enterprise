import { ListPolicyValidationsRequest } from '../../cluster-services/cluster_services.pb';
import { useListPolicyValidations } from '../../contexts/PolicyViolations';
import LoadingWrapper from '../Workspaces/WorkspaceDetails/Tabs/WorkspaceTabsWrapper';
import { FieldsType, PolicyViolationsTable } from './Table';

const PoliciesViolations = (req: ListPolicyValidationsRequest) => {
  const { data, error, isLoading } = useListPolicyValidations(req);
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
