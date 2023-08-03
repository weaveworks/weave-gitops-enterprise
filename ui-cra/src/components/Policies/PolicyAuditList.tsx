import { Alert } from '@material-ui/lab';
import { Flex, FluxObject } from '@weaveworks/weave-gitops';
import { useHistory } from 'react-router-dom';
import { useQueryService } from '../../hooks/query';
import { URLQueryStateManager } from '../Explorer/QueryStateManager';
import { QueryStateProvider } from '../Explorer/hooks';

// @ts-ignore
import PaginationControls from '../Explorer/PaginationControls';
import { useEffect, useState } from 'react';

const PolicyAuditList = () => {
  const [violations, setViolations] = useState<any[]>([]);
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
useEffect(()=>{

})
  const rows = data?.objects?.map(obj => {
    const details = JSON.parse(obj.unstructured || '');

    setViolations({ ...obj, ...details });
    // new FluxObject({ payload: obj.unstructured || '' });
    console.log(violations);
  });

  // const fields: Field[] = [
  //   {
  //     label: "Message",
  //     value:"message",
  //     textSearchable: true,
  //     sortValue: ({ message }) => message,
  //     maxWidth: 300,
  //   },
  //   ...(isFlagEnabled("WEAVE_GITOPS_FEATURE_CLUSTER")
  //     ? [
  //         {
  //           label: "Cluster",
  //           value: "clusterName",
  //           sortValue: ({ clusterName }) => clusterName,
  //         },
  //       ]
  //     : []),
  //   ...(!req.kind || req.kind === Kind.Policy
  //     ? [
  //         {
  //           label: "Application",
  //           value: ({ namespace, entity }) => `${namespace}/${entity}`,
  //           sortValue: ({ namespace, entity }) => `${namespace}/${entity}`,
  //         },
  //       ]
  //     : []),
  //   {
  //     label: "Severity",
  //     value: ({ severity }) => <Severity severity={severity || ""} />,
  //     sortValue: ({ severity }) => severity,
  //   },
  //   {
  //     label: "Category",
  //     value: "category",
  //     sortValue: ({ category }) => category,
  //   },
  //   ...(!req.kind || req.kind !== Kind.Policy
  //     ? [
  //         {
  //           label: "Violated Policy",
  //           value: "name",
  //           sortValue: ({ name }) => name,
  //         },
  //       ]
  //     : []),
  //   {
  //     label: "Violation Time",
  //     value: ({ createdAt }) => <Timestamp time={createdAt} />,
  //     defaultSort: true,
  //     sortValue: ({ createdAt }) => {
  //       const t = createdAt && new Date(createdAt).getTime();
  //       return t * -1;
  //     },
  //   },
  // ];
  return (
    <QueryStateProvider manager={manager}>
      <div>
        {error && <Alert severity="error">{error.message}</Alert>}
        <Flex wide></Flex>

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
