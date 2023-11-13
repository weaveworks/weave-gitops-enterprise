import { useFeatureFlags } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import { useContext } from 'react';
import { useQuery } from 'react-query';
import {
  DoQueryResponse,
  EnabledComponent,
  ListEnabledComponentsResponse,
} from '../api/query/query.pb';
import { QueryServiceContext } from '../contexts/QueryService';

type QueryOpts = {
  terms?: string;
  filters?: string[];
  limit?: number;
  offset?: number;
  orderBy?: string;
  category?: string;
  descending?: boolean;
};

// We want to "OR" filters of the same field, but "AND" filters of different fields.
// This function takes a list of filters and
// re-formats them to accomplish this against the query service backend.
export function formatFilters(filters: string[]) {
  const commonKinds: { [key: string]: string[] } = {};

  _.each(filters, filter => {
    const [kind, value] = filter.split(':');
    if (!commonKinds[kind]) {
      commonKinds[kind] = [];
    }
    commonKinds[kind].push(value);
  });

  const clauses: string[] = [];

  _.each(commonKinds, (values, kind) => {
    if (values.length > 1) {
      clauses.push(`${kind}:/(${values.join('|')})/`);
      return;
    }

    clauses.push(`${kind}:${values[0]}`);
  });

  return clauses;
}

export function useQueryService({
  terms,
  filters,
  limit,
  offset,
  orderBy,
  category,
  descending,
}: QueryOpts) {
  const api = useContext(QueryServiceContext);

  let formatted = formatFilters(filters || []);

  if (category) {
    formatted = _.concat(formatted, ['category:' + category]);
  }

  return useQuery<DoQueryResponse, Error>(
    ['query', { terms, filters, limit, offset, orderBy, descending }],
    () => {
      return api.DoQuery({
        terms,
        filters: formatted,
        limit,
        offset,
        orderBy,
        descending,
      });
    },
    {
      keepPreviousData: true,
      retry: false,
      refetchInterval: 10000,
    },
  );
}

export function useListAccessRules() {
  const api = useContext(QueryServiceContext);

  return useQuery(['listAccessRules'], () => api.DebugGetAccessRules({}));
}

export function useListFacets() {
  const api = useContext(QueryServiceContext);

  return useQuery(['facets'], () => api.ListFacets({}), {
    refetchIntervalInBackground: true,
    refetchInterval: 10000,
  });
}

function useListEnabledComponents() {
  const api = useContext(QueryServiceContext);

  const { isFlagEnabled } = useFeatureFlags();
  const isExplorerEnabled = isFlagEnabled('WEAVE_GITOPS_FEATURE_EXPLORER');

  return useQuery<ListEnabledComponentsResponse, Error>(
    ['enabledComponents', isExplorerEnabled],
    () => api.ListEnabledComponents({}),
    { enabled: isExplorerEnabled },
  );
}

export function useIsEnabledForComponent(cmp: EnabledComponent) {
  const { data } = useListEnabledComponents();

  return data?.components?.includes(cmp);
}
