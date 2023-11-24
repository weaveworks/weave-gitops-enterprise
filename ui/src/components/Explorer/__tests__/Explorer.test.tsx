import {
  act,
  queryByLabelText,
  render,
  RenderResult,
} from '@testing-library/react';
import {
  defaultContexts,
  MockQueryService,
  withContext,
} from '../../../utils/test-utils';
import Explorer from '../Explorer';
import { addFieldsWithIndex } from '../ExplorerTable';
import { APIContext } from '../../../contexts/API';

describe('addExplorerFields', () => {
  const newField = (id: string, index?: number) => ({
    id,
    label: id,
    value: id,
    index,
  });

  it('adds items to the end of the list if no index', () => {
    const fields = [newField('foo')];
    const fieldsToAdd = [newField('bar')];
    const newFields = addFieldsWithIndex(fields, fieldsToAdd);

    expect(newFields.map(field => field.id)).toEqual(['foo', 'bar']);
  });

  it('adds items to the beginning of the list if index is 0', () => {
    const fields = [newField('foo')];
    const fieldsToAdd = [newField('bar', 0)];
    const newFields = addFieldsWithIndex(fields, fieldsToAdd);
    expect(newFields.map(field => field.id)).toEqual(['bar', 'foo']);
  });

  it('adds item to middle of list if index is specified', () => {
    const fields = [newField('foo'), newField('baz')];
    const fieldsToAdd = [newField('bar', 1)];
    const newFields = addFieldsWithIndex(fields, fieldsToAdd);
    expect(newFields.map(field => field.id)).toEqual(['foo', 'bar', 'baz']);
  });
});

describe('Explorer', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: MockQueryService;

  beforeEach(() => {
    api = new MockQueryService();
    wrap = withContext([
      ...defaultContexts(),
      [APIContext.Provider, { value: { query: api } }],
    ]);
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
        orderDescending: false,
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
    //  @eslint-disable-next-line
    const input = queryByLabelText(
      container,
      'Kustomization',
    ) as HTMLInputElement;

    expect(input?.checked).toBeTruthy();
  });
  it('you can configure the visible columns', async () => {
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

    const tableFields = [
      {
        id: 'my-cool-column',
        label: 'My Cool Column',
        value: (o: any) => `${o.kind}-foo-bar`,
      },
    ];

    let result = {} as RenderResult;
    await act(async () => {
      const c = wrap(<Explorer fields={tableFields} />);
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
