import { act, fireEvent, render, screen } from '@testing-library/react';
import { ToggleSuspendTerraformObjectsResponse } from '../../../api/terraform/terraform.pb';
import { TerraformProvider } from '../../../contexts/Terraform';
import {
  TerraformClientMock,
  defaultContexts,
  withContext,
} from '../../../utils/test-utils';
import TerraformObjectDetail from '../TerraformObjectDetail';

const res = {
  object: {
    name: 'helloworld',
    namespace: 'flux-system',
    clusterName: 'management',
    sourceRef: {
      apiVersion: '',
      kind: 'GitRepository',
      name: 'helloworld',
      namespace: 'flux-system',
    },
    appliedRevision: '',
    path: './',
    interval: { hours: '0', minutes: '1', seconds: '0' },
    lastUpdatedAt: '',
    driftDetectionResult: false,
    inventory: [],
    conditions: [],
    suspended: false,
  },
  yaml: '',
  type: 'Terraform',
};

describe('TerraformObjectDetail', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: TerraformClientMock;

  beforeEach(() => {
    api = new TerraformClientMock();
    wrap = withContext([...defaultContexts(), [TerraformProvider, { api }]]);
  });

  it('syncs a terraform object', async () => {
    const params: any = res.object;
    api.GetTerraformObjectReturns = res;
    const recorder = jest.fn();

    api.SyncTerraformObjects = (...args) => {
      recorder(...args);
      return new Promise(() => ({}));
    };

    await act(async () => {
      const c = wrap(
        <TerraformObjectDetail
          name={params.name}
          namespace={params.namespace}
          clusterName="Default"
        />,
      );
      render(c);
    });

    const button = await screen.findByText('Sync');

    fireEvent.click(button);

    expect(recorder).toHaveBeenCalledWith({
      objects: [
        {
          name: 'helloworld',
          namespace: 'flux-system',
          clusterName: 'Default',
        },
      ],
    });
  });
  it('suspends a terraform object', async () => {
    const params: any = res.object;
    api.GetTerraformObjectReturns = res;
    const recorder = jest.fn();
    const p = new Promise<ToggleSuspendTerraformObjectsResponse>(() => ({}));

    api.ToggleSuspendTerraformObjects = (...args) => {
      recorder(...args);
      return p;
    };

    await act(async () => {
      const c = wrap(
        <TerraformObjectDetail
          name={params.name}
          namespace={params.namespace}
          clusterName="Default"
        />,
      );
      render(c);
    });

    const info = await screen.findByText('Suspended:');

    const suspendedValue = info?.parentNode?.textContent;

    expect(suspendedValue).toEqual('Suspended:False');

    const button = await screen.findByText('Suspend');

    fireEvent.click(button);

    expect(recorder).toHaveBeenCalledWith({
      objects: [
        {
          name: 'helloworld',
          namespace: 'flux-system',
          clusterName: 'Default',
        },
      ],
      suspend: true,
    });
  });
  it('calls get terraform object plan', async () => {
    const params: any = res.object;
    api.GetTerraformObjectReturns = res;
    const recorder = jest.fn();

    api.GetTerraformObjectPlan = (...args) => {
      recorder(...args);
      return new Promise(() => ({}));
    };

    await act(async () => {
      const c = wrap(
        <TerraformObjectDetail
          name={params.name}
          namespace={params.namespace}
          clusterName="Default"
        />,
      );
      render(c);
    });

    const tab = await screen.findByText('Plan');

    fireEvent.click(tab);

    expect(recorder).toHaveBeenCalledWith({
      name: 'helloworld',
      namespace: 'flux-system',
      clusterName: 'Default',
    });
  });
  it('calls replan terraform object', async () => {
    const params: any = res.object;
    api.GetTerraformObjectReturns = res;
    const recorder = jest.fn();

    api.ReplanTerraformObject = (...args) => {
      recorder(...args);
      return new Promise(() => ({}));
    };

    await act(async () => {
      const c = wrap(
        <TerraformObjectDetail
          name={params.name}
          namespace={params.namespace}
          clusterName="Default"
        />,
      );
      render(c);
    });

    const tab = await screen.findByText('Plan');

    fireEvent.click(tab);

    const replanBtn = await screen.findByTestId('replan-btn');

    fireEvent.click(replanBtn);

    expect(recorder).toHaveBeenCalledWith({
      name: 'helloworld',
      namespace: 'flux-system',
      clusterName: 'Default',
    });
  });
});
