import { QueryState } from './hooks';
import _ from 'lodash';
import qs from 'query-string';
import { useHistory } from 'react-router-dom';

// We split this into an interface to allow for different implmentations in the future.
// The default implementation is URLQueryStateManager, which reads from the browser URL.
// That might not be the best to use in all situations, so the explorer is configuragble.
export interface QueryStateManager {
  read(): QueryState;
  write(state: QueryState): void;
}

const DEFAULT_LIMIT = 25;

export class URLQueryStateManager implements QueryStateManager {
  private history: ReturnType<typeof useHistory>;

  constructor(history: ReturnType<typeof useHistory>) {
    this.history = history;

    this.read = this.read.bind(this);
    this.write = this.write.bind(this);
  }

  read(): QueryState {
    const { terms, qFilters, limit, orderBy, ascending, offset } = qs.parse(
      this.history.location.search,
      { arrayFormat: 'comma' },
    );

    const l = parseInt(limit as string);
    const o = parseInt(offset as string);

    return {
      terms: (terms as string) || '',
      limit: !_.isNaN(l) ? l : DEFAULT_LIMIT,
      offset: !_.isNaN(o) ? o : 0,
      filters: qFilters ? _.split(qFilters as string, ',') : [],
      orderBy: (orderBy as string) || '',
      orderAscending: ascending === 'true',
    };
  }

  write(queryState: QueryState): void {
    let offset: any = queryState.offset;
    let limit: any = queryState.limit;
    if (queryState.offset === 0) {
      offset = null;
      limit = null;
    }

    const q = qs.stringify(
      {
        terms: queryState.terms,
        qFilters: queryState.filters,
        limit,
        offset,
        ascending: queryState.orderAscending,
        orderBy: queryState.orderBy,
      },
      { skipNull: true, skipEmptyString: true, arrayFormat: 'comma' },
    );

    this.history.push(`?${q}`);
  }
}
