import { PolicyTable, RequestStateHandler } from '@weaveworks/weave-gitops';
import { useListPolicies } from '../../contexts/PolicyViolations';
import { TableWrapper } from '../Shared';
import { RequestError } from '../../types/custom';

export const PoliciesTab = () => {
  const { data, isLoading } = useListPolicies({});

  return (
    <RequestStateHandler
      loading={isLoading}
      error={data?.errors![0] as RequestError}
    >
      {data?.policies && (
        <TableWrapper id="policy-list">
          <PolicyTable policies={data.policies} />
        </TableWrapper>
      )}
    </RequestStateHandler>
  );
};
