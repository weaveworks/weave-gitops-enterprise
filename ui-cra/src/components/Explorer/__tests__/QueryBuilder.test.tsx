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
    const onPin = jest.fn();
    const onFilterSelect = jest.fn();

    const { rerender } = render(
      wrap(
        <QueryBuilder
          query=""
          pinnedTerms={[]}
          filters={[]}
          selectedFilter=""
          onChange={onChange}
          onPin={onPin}
          onFilterSelect={onFilterSelect}
        />,
      ),
    );

    let input = document.querySelector('input');

    act(() => {
      fireEvent.change(input as Element, { target: { value: 'test' } });
    });

    expect(onChange).toHaveBeenCalledWith('test', []);

    rerender(
      wrap(
        <QueryBuilder
          query={onChange.mock.calls[0][0]}
          pinnedTerms={[]}
          filters={[]}
          selectedFilter=""
          onChange={onChange}
          onPin={onPin}
          onFilterSelect={onFilterSelect}
        />,
      ),
    );

    input = document.querySelector('input');

    act(() => {
      fireEvent.keyDown(input as Element, { key: 'Enter', code: 'Enter' });
    });

    expect(onPin).toHaveBeenCalledWith(['test']);
  });
  it('selects filters', () => {
    const onChange = jest.fn();
    const onPin = jest.fn();
    const onFilterSelect = jest.fn();

    const c = render(
      wrap(
        <QueryBuilder
          query=""
          pinnedTerms={[]}
          filters={[{ value: 'name:Kustomization', label: 'Kustomizations' }]}
          selectedFilter=""
          onChange={onChange}
          onPin={onPin}
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

    c.rerender(
      wrap(
        <QueryBuilder
          query=""
          pinnedTerms={['kind:Kustomization']}
          filters={[{ value: 'kind:Kustomization', label: 'Kustomizations' }]}
          selectedFilter="kind:Kustomization"
          onChange={onChange}
          onPin={onPin}
          onFilterSelect={onFilterSelect}
        />,
      ),
    );

    const chips = document.querySelector('.MuiChip-root');

    expect(chips).toHaveTextContent('kind:Kustomization');
  });
});
