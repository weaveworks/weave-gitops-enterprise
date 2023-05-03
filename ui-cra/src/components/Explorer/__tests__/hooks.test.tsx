import { act, renderHook } from '@testing-library/react-hooks';
import { MemoryRouter } from 'react-router-dom';
import { filterChangeHandler, useQueryState } from '../hooks';

describe('useQueryState', () => {
  it('returns initial state', () => {
    const wrapper = ({ children }: any) => (
      <MemoryRouter>{children}</MemoryRouter>
    );
    const { result } = renderHook(
      () =>
        useQueryState({
          enableURLState: false,
          filters: [{ label: 'kind', value: 'kind:foo' }],
        }),
      { wrapper },
    );

    const initial = result.current[0];

    expect(initial).toEqual({
      filters: [{ label: 'kind', value: 'kind:foo' }],
      limit: 25,
      offset: 0,
      orderBy: '',
      orderAscending: false,
      query: '',
      selectedFilter: '',
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
          filters: [{ label: 'kind', value: 'kind:foo' }],
        }),
      { wrapper },
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
      orderBy: '',
      orderAscending: false,
      query: 'kind:bar',
      selectedFilter: 'kind:bar',
    });
  });
});
