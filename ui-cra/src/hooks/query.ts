import _ from 'lodash';
import { useQuery } from 'react-query';
import { Query, QueryClause, QueryResponse } from '../api/query/query.pb';

enum QueryOperands {
  equal = 'equal',
  notEqual = 'not_equal',
}

enum GlobalOperand {
  and = 'and',
  or = 'or',
}

const QUERY_REGEXP_STRING = '(.+?):(.+)';

function isSimpleQuery(clauses: string[]) {
  const queryRe = new RegExp(QUERY_REGEXP_STRING);

  if (clauses.length !== 1) {
    return false;
  }

  const matches = queryRe.exec(clauses[0]);

  if (!matches) {
    return true;
  }

  return false;
}

function convertToOpts(query: string): {
  clauses: QueryClause[];
  globalOperand: GlobalOperand;
} {
  if (!query) {
    return {
      clauses: [{ key: '', value: '' }],
      globalOperand: GlobalOperand.and,
    };
  }

  const clauses = query.split(',');

  if (isSimpleQuery(clauses)) {
    const value = clauses[0];

    return {
      clauses: [
        {
          key: 'name',
          value,
          operand: QueryOperands.equal,
        },
        {
          key: 'namespace',
          value,
          operand: QueryOperands.equal,
        },
        {
          key: 'cluster',
          value,
          operand: QueryOperands.equal,
        },
      ],
      globalOperand: GlobalOperand.or,
    };
  }

  const out = _.map(clauses, clause => {
    const queryRe = new RegExp(QUERY_REGEXP_STRING);
    const matches = queryRe.exec(clause);

    if (!matches) {
      throw new Error(
        'Invalid query. Only single-term or key:value queries are supported; single-term and key:value queries cannot be mixed.',
      );
    }

    return {
      key: matches[1],
      value: matches[2],
      operand: QueryOperands.equal,
    };
  });

  return { clauses: out, globalOperand: GlobalOperand.and };
}

export function useQueryService(
  query: string,
  limit: number = 2,
  offset: number = 0,
  orderBy?: string,
) {
  const api = Query;

  return useQuery<QueryResponse, Error>(
    ['query', { query, limit, offset, orderBy }],
    () => {
      const { clauses: q, globalOperand } = convertToOpts(query);

      return api.DoQuery({
        query: q,
        limit,
        offset,
        orderBy,
        globalOperand,
      });
    },
    {
      keepPreviousData: true,
      retry: false,
      refetchInterval: 5000,
    },
  );
}

export function useListAccessRules() {
  const api = Query;

  return useQuery(['listAccessRules'], () => api.DebugGetAccessRules({}));
}
