import { renderHook } from '@testing-library/react-hooks';
import QueryServiceProvider from '../../contexts/QueryService';

import { QueryClient, QueryClientProvider } from 'react-query';
import { MockQueryService } from '../../utils/test-utils';
import { formatFilters, useQueryService } from '../query';

describe('useQueryService', () => {
  let mock: MockQueryService;
  let wrapper: ({ children }: any) => JSX.Element;

  beforeEach(() => {
    mock = new MockQueryService();

    wrapper = ({ children }: any) => (
      <QueryClientProvider client={new QueryClient()}>
        {/* @ts-ignore */}
        <QueryServiceProvider api={mock}>{children}</QueryServiceProvider>
      </QueryClientProvider>
    );
  });
  it('does an OR within a filter field', () => {
    mock.DoQueryReturns = {
      objects: [],
    };

    mock.DoQuery = jest.fn();

    const filters = ['kind:Kustomization', 'kind:HelmRelease'];

    renderHook(() => useQueryService({ filters }), { wrapper });

    expect(mock.DoQuery).toHaveBeenCalledWith({
      filters: ['kind:(Kustomization|HelmRelease)'],
    });
  });
  it('does an AND between filter fields', () => {
    mock.DoQueryReturns = {
      objects: [],
    };

    mock.DoQuery = jest.fn();

    const filters = ['kind:Kustomization', 'cluster:management'];

    renderHook(() => useQueryService({ filters }), { wrapper });

    expect(mock.DoQuery).toHaveBeenCalledWith({
      filters: ['+kind:Kustomization', '+cluster:management'],
    });
  });
  describe('formatFilters', () => {
    it('combines filters with the same field', () => {
      const filters = ['kind:Kustomization', 'kind:HelmRelease'];

      const result = formatFilters(filters);

      expect(result).toEqual(['kind:(Kustomization|HelmRelease)']);
    });
    it('does not wrap parens around a single filter', () => {
      const filters = ['kind:Kustomization'];

      const result = formatFilters(filters);

      expect(result).toEqual(['+kind:Kustomization']);
    });
  });
});
