import _ from 'lodash';
import { useQuery } from 'react-query';
import { Query, QueryClause, QueryResponse } from '../api/query/query.pb';

enum QueryOperands {
  equal = 'equal',
  notEqual = 'not_equal',
}

function convertToOpts(query: string): QueryClause[] {
  if (!query) {
    return [{ key: '', value: '' }];
  }

  const clauses = query.split(',');

  const out = _.map(clauses, clause => {
    const queryRe = /(.+?):(.+)/g;
    const matches = queryRe.exec(clause);

    if (!matches) {
      throw new Error('Invalid query');
    }

    return {
      key: matches[1],
      value: matches[2],
      operand: QueryOperands.equal,
    };
  });

  return out;
}

export function useQueryService(
  query: string,
  limit: number = 2,
  offset: number = 0,
) {
  const api = Query;

  return useQuery<QueryResponse, Error>(
    ['query', { query, limit, offset }],
    () => {
      const opts = convertToOpts(query);

      return api.DoQuery({
        query: opts,
        limit,
        offset,
      });
    },
    {
      keepPreviousData: true,
      retry: false,
    },
  );
}

export function useListAccessRules() {
  const api = Query;

  return useQuery(['listAccessRules'], () => api.DebugGetAccessRules({}));
}
