import { act, renderHook } from '@testing-library/react-hooks';
import { filterChangeHandler, useQueryState } from '../hooks';

describe('useQueryState', () => {
  it('returns initial state', () => {
    const { result } = renderHook(() =>
      useQueryState({
        enableURLState: false,
        filters: [{ label: 'kind', value: 'kind:foo' }],
      }),
    );

    const initial = result.current[0];

    expect(initial).toEqual({
      filters: [{ label: 'kind', value: 'kind:foo' }],
      limit: 25,
      offset: 0,
      orderBy: 'name',
      orderDescending: false,
      pinnedTerms: [],
      query: '',
      selectedFilter: '',
    });
  });
  it('pins terms', () => {
    const { result } = renderHook(() =>
      useQueryState({
        enableURLState: false,
        filters: [{ label: 'kind', value: 'kind:foo' }],
      }),
    );

    act(() => {
      const setState = result.current[1];
      setState({
        ...result.current[0],
        query: 'test',
      });
    });

    const state = result.current[0];
    expect(state).toEqual({
      filters: [{ label: 'kind', value: 'kind:foo' }],
      limit: 25,
      offset: 0,
      orderBy: 'name',
      orderDescending: false,
      pinnedTerms: [],
      query: 'test',
      selectedFilter: '',
    });
  });
  it('filterChangeHandler', () => {
    const { result } = renderHook(() =>
      useQueryState({
        enableURLState: false,
        filters: [{ label: 'kind', value: 'kind:foo' }],
      }),
    );

    const handler = filterChangeHandler(result.current[0], result.current[1]);

    act(() => {
      handler('kind:bar');
    });

    const state = result.current[0];
    expect(state).toEqual({
      filters: [{ label: 'kind', value: 'kind:foo' }],
      limit: 25,
      offset: 0,
      orderBy: 'name',
      orderDescending: false,
      pinnedTerms: ['kind:bar'],
      query: '',
      selectedFilter: 'kind:bar',
    });
  });
});
