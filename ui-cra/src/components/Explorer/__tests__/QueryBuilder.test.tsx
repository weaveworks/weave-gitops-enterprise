import { act, fireEvent, render } from '@testing-library/react';
import { defaultContexts, withContext } from '../../../utils/test-utils';
import QueryBuilder from '../QueryBuilder';

describe('QueryBuilder', () => {
  let wrap: (el: JSX.Element) => JSX.Element;

  beforeEach(() => {
    wrap = withContext([...defaultContexts()]);
  });
  it('returns pinned terms', () => {
    const onChange = jest.fn();
    const onSubmit = jest.fn();
    const onFilterSelect = jest.fn();

    const { rerender } = render(
      wrap(
        <QueryBuilder
          query=""
          filters={[]}
          selectedFilter=""
          onChange={onChange}
          onSubmit={onSubmit}
          onFilterSelect={onFilterSelect}
        />,
      ),
    );

    let input = document.querySelector('input');

    act(() => {
      fireEvent.change(input as Element, { target: { value: 'test' } });
    });

    expect(onChange).toHaveBeenCalledWith('test');

    rerender(
      wrap(
        <QueryBuilder
          query={onChange.mock.calls[0][0]}
          filters={[]}
          selectedFilter=""
          onChange={onChange}
          onSubmit={onSubmit}
          onFilterSelect={onFilterSelect}
        />,
      ),
    );

    const form = document.querySelector('form');

    act(() => {
      fireEvent.submit(form as Element);
    });

    expect(onSubmit).toHaveBeenCalledWith('test');
  });
  it('selects filters', () => {
    const onChange = jest.fn();
    const onSubmit = jest.fn();
    const onFilterSelect = jest.fn();

    const c = render(
      wrap(
        <QueryBuilder
          query=""
          filters={[{ value: 'name:Kustomization', label: 'Kustomizations' }]}
          selectedFilter=""
          onChange={onChange}
          onSubmit={onSubmit}
          onFilterSelect={onFilterSelect}
        />,
      ),
    );

    const select = c.getByPlaceholderText('Filters');

    act(() => {
      fireEvent.change(select as Element, {
        target: { value: 'name:Kustomization' },
      });
    });

    expect(onFilterSelect).toHaveBeenCalledWith('name:Kustomization');
  });
});
