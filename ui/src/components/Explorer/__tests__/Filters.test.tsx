import { act, fireEvent, render, screen } from '@testing-library/react';
import { withContext } from '../../../utils/test-utils';
import Filters from '../Filters';
import { QueryState, QueryStateProvider } from '../hooks';
import { QueryStateManager } from '../QueryStateManager';

describe('Filters', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let manager: QueryStateManager;

  beforeEach(() => {
    manager = {
      read: jest.fn(() => ({
        terms: '',
        filters: [],
        limit: 0,
        offset: 0,
        orderBy: '',
        orderAscending: false,
      })),
      write: jest.fn(),
    };

    wrap = withContext([[QueryStateProvider, { manager }]]);
  });

  it('selects filters', () => {
    const facets = [
      {
        field: 'Kind',
        values: [
          'Kustomization',
          'HelmRelease',
          'GitRepository',
          'HelmRepository',
          'Bucket',
          'HelmChart',
          'OCIRepository',
        ],
      },
      {
        field: 'Namespace',
        values: ['default', 'flux-system', 'flux'],
      },
    ];

    const c = wrap(<Filters facets={facets} />);
    const { rerender } = render(c);
    const qs: QueryState = {
      filters: [],
      terms: '',
      limit: 0,
      offset: 0,
      orderBy: '',
      orderAscending: false,
    };

    manager.read = jest.fn(() => qs);

    const input1 = screen.queryByLabelText('Kustomization') as HTMLInputElement;

    expect(input1?.checked).toBeFalsy();

    act(() => {
      fireEvent.click(screen.getByText('Kustomization'));
    });

    expect(manager.write).toHaveBeenLastCalledWith({
      ...qs,
      filters: ['Kind:Kustomization'],
    });

    manager.read = jest.fn(() => ({
      ...qs,
      filters: ['Kind:Kustomization'],
    }));

    rerender(wrap(<Filters facets={facets} />));

    const input2 = screen.queryByLabelText('Kustomization') as HTMLInputElement;

    expect(input2?.checked).toBeTruthy();

    act(() => {
      fireEvent.click(screen.getByText('HelmRelease'));
    });

    expect(manager.write).toHaveBeenLastCalledWith({
      ...qs,
      filters: ['Kind:Kustomization', 'Kind:HelmRelease'],
    });

    manager.read = jest.fn(() => ({
      ...qs,
      filters: ['Kind:Kustomization', 'Kind:HelmRelease'],
    }));

    rerender(wrap(<Filters facets={facets} />));

    const input3 = screen.queryByLabelText('HelmRelease') as HTMLInputElement;

    expect(input3?.checked).toBeTruthy();

    act(() => {
      fireEvent.click(screen.getByText('HelmRelease'));
    });

    rerender(wrap(<Filters facets={facets} />));

    // Make sure something gets removed if its clicked again
    expect(manager.write).toHaveBeenLastCalledWith({
      ...qs,
      filters: ['Kind:Kustomization'],
    });
  });
});
