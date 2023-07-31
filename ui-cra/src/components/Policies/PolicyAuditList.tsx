import { Alert } from '@material-ui/lab';
import { Flex, FluxObject } from '@weaveworks/weave-gitops';
import { useHistory } from 'react-router-dom';
import { useQueryService } from '../../hooks/query';
import { URLQueryStateManager } from '../Explorer/QueryStateManager';
import { QueryStateProvider } from '../Explorer/hooks';

// @ts-ignore
import PaginationControls from '../Explorer/PaginationControls';

const PolicyAuditList = () => {
  const history = useHistory();
  const manager = new URLQueryStateManager(history);

  const queryState = manager.read();
  const setQueryState = manager.write;

  const { data, error } = useQueryService({
    terms: queryState.terms,
    filters: ['kind:Event', ...queryState.filters],
    limit: queryState.limit,
    offset: queryState.offset,
    orderBy: queryState.orderBy,
    ascending: queryState.orderAscending,
  });

 const rows= data?.objects?.map(obj => {
    const fObj = new FluxObject({ payload: obj.unstructured || '' });
    return {
      ...fObj.obj,
      ...obj,
    };
 });
  console.log(rows)
  return (
    <QueryStateProvider manager={manager}>
      <div>
        {error && <Alert severity="error">{error.message}</Alert>}
        <Flex wide>

        </Flex>

        <PaginationControls
          queryState={queryState}
          setQueryState={setQueryState}
          count={data?.objects?.length || 0}
        />
      </div>
    </QueryStateProvider>
  );
};

export default PolicyAuditList;
