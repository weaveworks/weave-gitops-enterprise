import { PolicyTable } from '@weaveworks/weave-gitops';
import { useListPolicies } from '../../contexts/PolicyViolations';
import { TableWrapper } from '../Shared';
import LoadingWrapper from '../Workspaces/WorkspaceDetails/Tabs/WorkspaceTabsWrapper';

export const PoliciesTab = () => {
  const { data, isLoading } = useListPolicies({});

  return (
    <LoadingWrapper loading={isLoading} errors={data?.errors}>
      {data?.policies && (
        <TableWrapper id="policy-list">
          <PolicyTable policies={data.policies} />
        </TableWrapper>
      )}
    </LoadingWrapper>
  );
};
