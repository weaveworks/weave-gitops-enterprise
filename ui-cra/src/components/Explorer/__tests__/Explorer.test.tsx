import {
  RenderResult,
  act,
  queryByLabelText,
  render,
} from '@testing-library/react';
import QueryServiceProvider from '../../../contexts/QueryService';
import {
  MockQueryService,
  defaultContexts,
  withContext,
} from '../../../utils/test-utils';
import Explorer from '../Explorer';

describe('Explorer', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: MockQueryService;

  beforeEach(() => {
    api = new MockQueryService();
    wrap = withContext([...defaultContexts(), [QueryServiceProvider, { api }]]);
  });

  it('renders rows', async () => {
    const objects = [
      {
        kind: 'Kustomization',
        name: 'flux-system',
        namespace: 'flux-system',
        status: 'Ready',
      },
      {
        kind: 'HelmRelease',
        name: 'flux-system',
        namespace: 'flux-system',
        status: 'Ready',
      },
    ];

    api.DoQueryReturns = {
      objects,
    };

    let result: RenderResult;
    await act(async () => {
      const c = wrap(<Explorer />);
      result = await render(c);
    });

    // @ts-ignore
    expect(result.container).toHaveTextContent(objects[0].name);
  });
  it('renders filter elements', async () => {
    api.DoQueryReturns = {
      objects: [],
    };

    const facets = [
      {
        field: 'Kind',
        values: ['Kustomization', 'HelmRelease', 'GitRepository'],
      },
    ];

    api.ListFacetsReturns = {
      facets,
    };

    let result: RenderResult;

    await act(async () => {
      const c = wrap(<Explorer />);
      result = await render(c);
    });

    // @ts-ignore
    expect(result.container).toHaveTextContent(facets[0].field);
    // @ts-ignore
    expect(result.container).toHaveTextContent(facets[0].values[0]);
  });
  it('renders filter state', async () => {
    api.DoQueryReturns = {
      objects: [],
    };

    const facets = [
      {
        field: 'Kind',
        values: ['Kustomization', 'HelmRelease', 'GitRepository'],
      },
    ];

    api.ListFacetsReturns = {
      facets,
    };

    const manager = {
      read: jest.fn(() => ({
        terms: '',
        filters: ['Kind:Kustomization'],
        limit: 0,
        offset: 0,
        orderBy: '',
        orderAscending: false,
      })),

      write: jest.fn(),
    };

    let result: RenderResult;

    await act(async () => {
      const c = wrap(<Explorer manager={manager} />);
      result = await render(c);
    });

    //  @ts-ignore
    const container = result.container;

    const input = queryByLabelText(
      container,
      'Kustomization',
    ) as HTMLInputElement;

    expect(input?.checked).toBeTruthy();
  });
  it('shows extra columns', async () => {
    const objects = [
      {
        kind: 'Kustomization',
        name: 'flux-system',
        namespace: 'flux-system',
        status: 'Ready',
      },
      {
        kind: 'HelmRelease',
        name: 'flux-system',
        namespace: 'flux-system',
        status: 'Ready',
      },
    ];

    api.DoQueryReturns = {
      objects,
    };

    const extraCols = [
      {
        label: 'My Cool Column',
        value: (o: any) => `${o.kind}-foo-bar`,
      },
    ];

    let result = {} as RenderResult;
    await act(async () => {
      const c = wrap(<Explorer extraColumns={extraCols} />);
      result = await render(c);
    });

    expect(result.container).toHaveTextContent('My Cool Column');
    expect(result.container).toHaveTextContent('Kustomization-foo-bar');
  });

  describe('snapshots', () => {
    it('renders', async () => {
      let result: RenderResult;

      api.DoQueryReturns = {
        objects: [
          {
            kind: 'Kustomization',
            name: 'flux-system',
            namespace: 'flux-system',
            status: 'Ready',
          },
        ],
      };
      api.ListFacetsReturns = {
        facets: [
          {
            field: 'Kind',
            values: ['Kustomization', 'HelmRelease', 'GitRepository'],
          },
        ],
      };

      await act(async () => {
        const c = wrap(<Explorer />);
        result = await render(c);
      });

      //   @ts-ignore
      expect(result.container).toMatchSnapshot();
    });
  });
});
