import {
  useFeatureFlags
} from '@weaveworks/weave-gitops';
import { useHistory } from 'react-router-dom';
import { useQueryService } from '../../../hooks/query';
import { URLQueryStateManager } from '../../Explorer/QueryStateManager';

// @ts-ignore
import { AuditTable } from './AuditTable';
import WarningMsg from './WarningMsg';

const PolicyAuditList = () => {
  const { isFlagEnabled } = useFeatureFlags();
  const useQueryServiceBackend = isFlagEnabled(
    'WEAVE_GITOPS_FEATURE_QUERY_SERVICE_BACKEND',
  );
  
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
    ascending: queryState.orderAscending,
  });

  return useQueryServiceBackend ? (
    <AuditTable
      data={data}
      error={error}
      isLoading={isLoading}
      queryState={queryState}
      setQueryState={setQueryState}
      manager={manager}
    />
  ) : (
    <WarningMsg />
  );
};
export default PolicyAuditList;
