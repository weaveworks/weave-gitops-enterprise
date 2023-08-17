import { RequestStateHandler, useFeatureFlags } from '@weaveworks/weave-gitops';
import { useHistory } from 'react-router-dom';
import { useQueryService } from '../../../hooks/query';
import { RequestError } from '../../../types/custom';
import { URLQueryStateManager } from '../../Explorer/QueryStateManager';
import { QueryStateProvider } from '../../Explorer/hooks';
import { AuditTable } from './AuditTable';
import WarningMsg from './WarningMsg';

const PolicyAuditList = () => {
  const history = useHistory();
  const { isFlagEnabled } = useFeatureFlags();
  const useQueryServiceBackend = isFlagEnabled(
    'WEAVE_GITOPS_FEATURE_QUERY_SERVICE_BACKEND',
  );
  const manager = new URLQueryStateManager(history);
  const queryState = manager.read();
  const setQueryState = manager.write;
  const { data, error, isLoading } = useQueryService({
    terms: queryState.terms,
    filters: ['kind:Event', ...queryState.filters],
    limit: queryState.limit,
    offset: queryState.offset,
    orderBy: queryState.orderBy,
    ascending: queryState.orderAscending,
  });

  return (
    <QueryStateProvider manager={manager}>
      <RequestStateHandler error={error as RequestError} loading={isLoading}>
        {useQueryServiceBackend ? (
          data?.objects?.length && (
            <AuditTable
              data={data}
              queryState={queryState}
              setQueryState={setQueryState}
            />
          )
        ) : (
          <WarningMsg />
        )}
      </RequestStateHandler>
    </QueryStateProvider>
  );
};
export default PolicyAuditList;
