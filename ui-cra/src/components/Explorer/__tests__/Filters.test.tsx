import { act, fireEvent, render } from '@testing-library/react';
import Filters from '../Filters';

describe('Filters', () => {
  it('selects filters', () => {
    const facets = [
      {
        field: 'kind',
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
        field: 'namespace',
        values: ['default', 'flux-system', 'flux'],
      },
    ];

    const state = {};

    const onFilterSelect = jest.fn();

    const { rerender, getByText, queryByLabelText } = render(
      <Filters facets={facets} state={state} onFilterSelect={onFilterSelect} />,
    );

    const input1 = queryByLabelText('Kustomization') as HTMLInputElement;

    expect(input1?.checked).toBeFalsy();

    act(() => {
      fireEvent.click(getByText('Kustomization'));
    });

    expect(onFilterSelect).toHaveBeenCalledWith({
      'kind:Kustomization': true,
    });

    rerender(
      <Filters
        facets={facets}
        state={{ 'kind:Kustomization': true }}
        onFilterSelect={onFilterSelect}
      />,
    );

    const input2 = queryByLabelText('Kustomization') as HTMLInputElement;

    expect(input2?.checked).toBeTruthy();

    act(() => {
      fireEvent.click(getByText('HelmRelease'));
    });

    expect(onFilterSelect).toHaveBeenCalledWith({
      'kind:Kustomization': true,
      'kind:HelmRelease': true,
    });

    rerender(
      <Filters
        facets={facets}
        state={{ 'kind:Kustomization': true, 'kind:HelmRelease': true }}
        onFilterSelect={onFilterSelect}
      />,
    );

    const input3 = queryByLabelText('HelmRelease') as HTMLInputElement;

    expect(input3?.checked).toBeTruthy();
  });
});
