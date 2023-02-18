import { act, fireEvent, render, screen } from '@testing-library/react';
import { CoreClientContextProvider, Kind } from '@weaveworks/weave-gitops';
import ImageAutomationPage from '..';
import {
  CoreClientMock,
  defaultContexts,
  TestFilterableTable,
  withContext,
} from '../../../utils/test-utils';

const IsCRDAvailableReturn = {
  clusters: {
    'default/tw-cluster-2': true,
    management: false,
  },
};

const imageRepositories = {
  objects: [
    {
      payload:
        '{"apiVersion":"image.toolkit.fluxcd.io/v1beta1","kind":"ImageRepository","metadata":{"creationTimestamp":"2023-02-14T13:09:41Z","finalizers":["finalizers.fluxcd.io"],"generation":1,"labels":{"kustomize.toolkit.fluxcd.io/name":"flux-system","kustomize.toolkit.fluxcd.io/namespace":"flux-system"},"managedFields":[{"apiVersion":"image.toolkit.fluxcd.io/v1beta1","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:labels":{"f:kustomize.toolkit.fluxcd.io/name":{},"f:kustomize.toolkit.fluxcd.io/namespace":{}}},"f:spec":{"f:image":{},"f:interval":{}}},"manager":"kustomize-controller","operation":"Apply","time":"2023-02-14T13:09:41Z"},{"apiVersion":"image.toolkit.fluxcd.io/v1beta1","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:finalizers":{".":{},"v:\\"finalizers.fluxcd.io\\"":{}}}},"manager":"image-reflector-controller","operation":"Update","time":"2023-02-14T13:09:41Z"},{"apiVersion":"image.toolkit.fluxcd.io/v1beta1","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:canonicalImageName":{},"f:conditions":{},"f:lastScanResult":{".":{},"f:scanTime":{},"f:tagCount":{}},"f:observedGeneration":{}}},"manager":"image-reflector-controller","operation":"Update","subresource":"status","time":"2023-02-14T13:09:42Z"}],"name":"howto-kubeconfig-dev","namespace":"default","resourceVersion":"2170198","uid":"3269380d-6597-4daa-83d2-36487bd360c7"},"spec":{"image":"img.hephy.pro/examples/howto-kubeconfig-dev","interval":"20s"},"status":{"canonicalImageName":"img.hephy.pro/examples/howto-kubeconfig-dev","conditions":[{"lastTransitionTime":"2023-02-14T13:09:42Z","message":"successful scan, found 48 tags","reason":"ReconciliationSucceeded","status":"True","type":"Ready"}],"lastScanResult":{"scanTime":"2023-02-15T15:08:30Z","tagCount":48},"observedGeneration":1}}\n',
      clusterName: 'management',
      tenant: '',
      uid: '3269380d-6597-4daa-83d2-36487bd360c7',
      inventory: [],
    },
    {
      payload:
        '{"apiVersion":"image.toolkit.fluxcd.io/v1beta1","kind":"ImageRepository","metadata":{"creationTimestamp":"2023-02-14T13:09:41Z","finalizers":["finalizers.fluxcd.io"],"generation":1,"labels":{"kustomize.toolkit.fluxcd.io/name":"flux-system","kustomize.toolkit.fluxcd.io/namespace":"flux-system"},"managedFields":[{"apiVersion":"image.toolkit.fluxcd.io/v1beta1","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:labels":{"f:kustomize.toolkit.fluxcd.io/name":{},"f:kustomize.toolkit.fluxcd.io/namespace":{}}},"f:spec":{"f:image":{},"f:interval":{}}},"manager":"kustomize-controller","operation":"Apply","time":"2023-02-14T13:09:41Z"},{"apiVersion":"image.toolkit.fluxcd.io/v1beta1","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:finalizers":{".":{},"v:\\"finalizers.fluxcd.io\\"":{}}}},"manager":"image-reflector-controller","operation":"Update","time":"2023-02-14T13:09:41Z"},{"apiVersion":"image.toolkit.fluxcd.io/v1beta1","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:canonicalImageName":{},"f:conditions":{},"f:lastScanResult":{".":{},"f:scanTime":{},"f:tagCount":{}},"f:observedGeneration":{}}},"manager":"image-reflector-controller","operation":"Update","subresource":"status","time":"2023-02-14T13:09:41Z"}],"name":"podinfo-repo","namespace":"default","resourceVersion":"2127616","uid":"c073419d-8b49-44cc-b526-27cdd8d44e2e"},"spec":{"image":"ghcr.io/stefanprodan/podinfo","interval":"1h"},"status":{"canonicalImageName":"ghcr.io/stefanprodan/podinfo","conditions":[{"lastTransitionTime":"2023-02-14T13:09:41Z","message":"successful scan, found 41 tags","reason":"ReconciliationSucceeded","status":"True","type":"Ready"}],"lastScanResult":{"scanTime":"2023-02-15T14:09:57Z","tagCount":41},"observedGeneration":1}}\n',
      clusterName: 'management',
      tenant: '',
      uid: 'c073419d-8b49-44cc-b526-27cdd8d44e2e',
      inventory: [],
    },
  ],
  errors: [],
};

describe('Image automation', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let core: CoreClientMock;
  beforeEach(() => {
    core = new CoreClientMock();
    wrap = withContext([
      ...defaultContexts(),
      [CoreClientContextProvider, { api: core }],
    ]);
  });
  it('renders onboarding message', async () => {
    core.IsCRDAvailableReturn = {
      'imageupdateautomations.image.toolkit.fluxcd.io': {
        clusters: { management: false },
      },
    };
    await act(async () => {
      const c = wrap(<ImageAutomationPage />);
      render(c);
    });
    const btn = document.getElementById('navigate-to-imageautomation');
    expect(btn?.textContent).toContain('IMAGE AUTOMATION GUIDE');
  });

  it('renders image automation page', async () => {
    core.IsCRDAvailableReturn = {
      'imageupdateautomations.image.toolkit.fluxcd.io': IsCRDAvailableReturn,
    };
    core.ListObjectsReturns = {
      [Kind.ImageRepository]: imageRepositories,
    };
    await act(async () => {
      const c = wrap(<ImageAutomationPage />);
      render(c);
    });

    expect(await screen.findByText('Image Automation')).toBeTruthy();
    const tabs = screen.getAllByRole('tab');
    expect(tabs).toHaveLength(3);
    expect(tabs[0]).toHaveTextContent('Image Repositories');
    expect(tabs[1]).toHaveTextContent('Image Policies');
    expect(tabs[2]).toHaveTextContent('Image Update Automations');

    //Test Image Repositories table view
    const filterTable = new TestFilterableTable(
      'image-repository-list',
      fireEvent,
    );

    filterTable.testRenderTable(
      ['Name', 'Namespace', 'Cluster Name', 'Status', 'Interval', 'Tag Count'],
      imageRepositories.objects.length,
    );
  });
});
