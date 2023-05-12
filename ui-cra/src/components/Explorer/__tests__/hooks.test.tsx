import { act, renderHook } from '@testing-library/react-hooks';
import { MemoryRouter } from 'react-router-dom';
import { QueryState } from '../hooks';
import { linkToExplorer } from '../utils';

describe('useQueryState', () => {
  it('returns initial state', () => {
    const wrapper = ({ children }: any) => (
      <MemoryRouter>{children}</MemoryRouter>
    );
    const { result } = renderHook(
      () =>
        useQueryState({
          enableURLState: false,
        }),
      { wrapper },
    );

    const initial = result.current[0];

    expect(initial).toEqual({
      filters: [],
      limit: 25,
      offset: 0,
      orderBy: '',
      orderAscending: false,
      terms: '',
    });
  });
  it('filterChangeHandler', () => {
    const wrapper = ({ children }: any) => (
      <MemoryRouter>{children}</MemoryRouter>
    );

    const { result } = renderHook(
      () =>
        useQueryState({
          enableURLState: false,
        }),
      { wrapper },
    );

    const handler = filterChangeHandler(result.current[0], result.current[1]);

    act(() => {
      handler({ 'kind:bar': true });
    });

    const state = result.current[0];
    expect(state).toEqual({
      filters: ['kind:bar'],
      limit: 25,
      offset: 0,
      orderBy: '',
      orderAscending: false,
      terms: '',
    });
  });

  describe('url state', () => {
    it('loads state with an irrelevant query', () => {
      const wrapper = ({ children }: any) => (
        <MemoryRouter initialEntries={['/explorer/query?foo=bar']}>
          {children}
        </MemoryRouter>
      );

      const { result } = renderHook(
        () =>
          useQueryState({
            enableURLState: true,
          }),
        { wrapper },
      );

      expect(result.current[0]).toEqual({
        filters: [],
        limit: 25,
        offset: 0,
        orderBy: '',
        orderAscending: false,
        terms: '',
      });
    });
    it('loads state with filters set', () => {
      const url = linkToExplorer(`/explorer/query`, {
        filters: ['kind:bar', 'kind:baz'],
      } as QueryState);
      const wrapper = ({ children }: any) => (
        <MemoryRouter initialEntries={[url]}>{children}</MemoryRouter>
      );

      const { result } = renderHook(
        () =>
          useQueryState({
            enableURLState: true,
          }),
        { wrapper },
      );

      expect(result.current[0]).toEqual({
        filters: ['kind:bar', 'kind:baz'],
        limit: 25,
        offset: 0,
        orderBy: '',
        orderAscending: false,
        terms: '',
      });
    });
  });
  describe('usePersistURL', () => {
    it('persists state to url', () => {
      const history = {
        replace: jest.fn(),
      };
      const wrapper = ({ children }: any) => (
        <MemoryRouter>{children}</MemoryRouter>
      );

      const queryState = {
        filters: ['kind:bar', 'kind:baz'],
      } as QueryState;

      renderHook(() => usePersistURL(history as any, queryState, true), {
        wrapper,
      });

      expect(history.replace).toHaveBeenCalledWith(
        '?qFilters=kind%3Abar,kind%3Abaz',
      );

      const queryState2 = {
        filters: ['kind:bar'],
      } as QueryState;

      renderHook(() => usePersistURL(history as any, queryState2, true), {
        wrapper,
      });

      expect(history.replace).toHaveBeenLastCalledWith('?qFilters=kind%3Abar');
    });
  });
});
