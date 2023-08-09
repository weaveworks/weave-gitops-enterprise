import { useFeatureFlags } from '@weaveworks/weave-gitops';
import { useHistory, useLocation } from 'react-router-dom';
import { useQueryService } from '../../../hooks/query';
import { URLQueryStateManager } from '../../Explorer/QueryStateManager';

// @ts-ignore
import { useEffect, useState } from 'react';
import LoadingWrapper from '../../Workspaces/WorkspaceDetails/Tabs/WorkspaceTabsWrapper';
import { AuditTable } from './AuditTable';
import WarningMsg from './WarningMsg';

const PolicyAuditList = () => {
  const [areQueryParamsRemoved, setAreQueryParamsRemoved] =
    useState<boolean>(false);
  const { isFlagEnabled } = useFeatureFlags();
  const useQueryServiceBackend = isFlagEnabled(
    'WEAVE_GITOPS_FEATURE_QUERY_SERVICE_BACKEND',
  );
  const history = useHistory();
  const location = useLocation();

  const manager = new URLQueryStateManager(history);
  const queryState = manager.read();
  const setQueryState = manager.write;

  useEffect(() => {
    const params = new URLSearchParams();
    params.delete('search');
    history.replace({
      ...location,
      search: params.toString(),
    });
    setAreQueryParamsRemoved(true);
  }, []);
  
  const { data, error, isLoading } = useQueryService({
    terms: queryState.terms,
    filters: ['kind:Event', ...queryState.filters],
    limit: queryState.limit,
    offset: queryState.offset,
    orderBy: queryState.orderBy,
    ascending: queryState.orderAscending,
  });

  return (
    <LoadingWrapper loading={!areQueryParamsRemoved}>
      {useQueryServiceBackend ? (
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
      )}
    </LoadingWrapper>
  );
};
export default PolicyAuditList;
