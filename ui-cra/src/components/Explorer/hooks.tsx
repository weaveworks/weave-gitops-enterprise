import { Field } from '@weaveworks/weave-gitops/ui/components/DataTable';
import _ from 'lodash';
import { createContext, useContext } from 'react';
import { QueryStateManager } from './QueryStateManager';

export type QueryState = {
  terms: string;
  filters: string[];
  limit: number;
  offset: number;
  orderBy: string;
  orderAscending: boolean;
};

export const columnHeaderHandler =
  (queryState: QueryState, setQueryState: (next: QueryState) => void) =>
  (field: Field) => {
    let col = _.isFunction(field.value)
      ? field.sortValue && field.sortValue(field.value)
      : field.value;

    // Override column name to maintain compatibility with the DataTable sync buttons
    if (col === 'clusterName') {
      col = 'cluster';
    }

    setQueryState({
      ...queryState,
      orderBy: col as string,
      orderAscending:
        queryState.orderBy === col ? !queryState.orderAscending : false,
    });
  };

const QueryStateManagerContext = createContext<QueryStateManager>(null as any);

export function QueryStateProvider({
  children,
  manager,
}: {
  children: React.ReactNode;
  manager: QueryStateManager;
}) {
  return (
    <QueryStateManagerContext.Provider value={manager}>
      {children}
    </QueryStateManagerContext.Provider>
  );
}

export function useSetQueryState() {
  const mgr = useContext(QueryStateManagerContext);

  return mgr.write;
}

export function useReadQueryState(): QueryState {
  const mgr = useContext(QueryStateManagerContext);

  return mgr.read();
}
