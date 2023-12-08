import { renderHook } from '@testing-library/react-hooks';
import { QueryClient, QueryClientProvider } from 'react-query';
import { Query } from '../../api/query/query.pb';
import { APIs, EnterpriseClientContext } from '../../contexts/API';
import { MockQueryService, newMockQueryService } from '../../utils/test-utils';
import { formatFilters, useQueryService } from '../query';

describe('useQueryService', () => {
  let mock: MockQueryService;
  let wrapper: ({ children }: any) => JSX.Element;

  beforeEach(() => {
    mock = newMockQueryService();

    wrapper = ({ children }: any) => {
      const query: Query = mock;
      return (
        <QueryClientProvider client={new QueryClient()}>
          <EnterpriseClientContext.Provider
            // ignore that we don't provide the other apis
            value={{ query } as unknown as APIs}
          >
            {children}
          </EnterpriseClientContext.Provider>
        </QueryClientProvider>
      );
    };
  });
  it('does an OR within a filter field', () => {
    mock.DoQueryReturns = {
      objects: [],
    };

    mock.DoQuery = jest.fn();

    const filters = ['kind:Kustomization', 'kind:HelmRelease'];

    renderHook(() => useQueryService({ filters }), { wrapper });

    expect(mock.DoQuery).toHaveBeenCalledWith({
      filters: ['kind:/(Kustomization|HelmRelease)/'],
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
      filters: ['kind:Kustomization', 'cluster:management'],
    });
  });
  describe('formatFilters', () => {
    it('combines filters with the same field', () => {
      const filters = ['kind:Kustomization', 'kind:HelmRelease'];

      const result = formatFilters(filters);

      expect(result).toEqual(['kind:/(Kustomization|HelmRelease)/']);
    });
    it('does not wrap parens around a single filter', () => {
      const filters = ['kind:Kustomization'];

      const result = formatFilters(filters);

      expect(result).toEqual(['kind:Kustomization']);
    });
  });
});
