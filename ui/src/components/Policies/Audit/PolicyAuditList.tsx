import { RequestStateHandler } from '@weaveworks/weave-gitops';
import { useHistory } from 'react-router-dom';
import { useQueryService } from '../../../hooks/query';
import { RequestError } from '../../../types/custom';
import { QueryStateProvider } from '../../Explorer/hooks';
import { URLQueryStateManager } from '../../Explorer/QueryStateManager';
import { AuditTable } from './AuditTable';

const PolicyAuditList = () => {
  const history = useHistory();
  const manager = new URLQueryStateManager(history);
  const queryState = manager.read();
  const setQueryState = manager.write;
  const { data, error, isLoading } = useQueryService({
    terms: queryState.terms,
    filters: ['kind:Event', ...queryState.filters],
    limit: queryState.limit,
    offset: queryState.offset,
    orderBy: queryState.orderBy,
    descending: queryState.orderDescending,
  });

  return (
    <QueryStateProvider manager={manager}>
      <RequestStateHandler error={error as RequestError} loading={isLoading}>
        <AuditTable
          objects={data?.objects || []}
          queryState={queryState}
          setQueryState={setQueryState}
        />
      </RequestStateHandler>
    </QueryStateProvider>
  );
};
export default PolicyAuditList;
