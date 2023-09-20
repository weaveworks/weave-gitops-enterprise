import { act, fireEvent, render, screen } from '@testing-library/react';
import ImageAutomationPage from '..';
import {
  CoreClientMock,
  defaultContexts,
  TestFilterableTable,
  withContext,
} from '../../../utils/test-utils';
import ImagePoliciesTable from '../policies/ImagePoliciesTable';
import ImageAutomationUpdatesTable from '../updates/ImageAutomationUpdatesTable';
import {
  CoreClientContextProvider,
  ImagePolicy,
  ImageRepository,
  ImageUpdateAutomation,
  Kind,
  showInterval,
} from '@weaveworks/weave-gitops';
import moment from 'moment';

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
const imagePolicies = {
  objects: [
    {
      payload:
        '{"apiVersion":"image.toolkit.fluxcd.io/v1beta1","kind":"ImagePolicy","metadata":{"creationTimestamp":"2023-02-06T14:26:01Z","finalizers":["finalizers.fluxcd.io"],"generation":2,"labels":{"kustomize.toolkit.fluxcd.io/name":"flux-system","kustomize.toolkit.fluxcd.io/namespace":"flux-system"},"managedFields":[{"apiVersion":"image.toolkit.fluxcd.io/v1beta1","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:labels":{"f:kustomize.toolkit.fluxcd.io/name":{},"f:kustomize.toolkit.fluxcd.io/namespace":{}}},"f:spec":{"f:imageRepositoryRef":{"f:name":{}},"f:policy":{"f:semver":{"f:range":{}}}}},"manager":"kustomize-controller","operation":"Apply","time":"2023-02-06T15:26:21Z"},{"apiVersion":"image.toolkit.fluxcd.io/v1beta1","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:finalizers":{".":{},"v:\\"finalizers.fluxcd.io\\"":{}}},"f:status":{"f:conditions":{},"f:latestImage":{},"f:observedGeneration":{}}},"manager":"image-reflector-controller","operation":"Update","time":"2023-02-06T15:26:22Z"}],"name":"podinfo","namespace":"flux-system","resourceVersion":"13621","uid":"5009e51d-0fee-4f8e-9df1-7684c8aac4bd"},"spec":{"imageRepositoryRef":{"name":"podinfo"},"policy":{"semver":{"range":"5.0.x"}}},"status":{"conditions":[{"lastTransitionTime":"2023-02-06T15:26:22Z","message":"Latest image tag for \'ghcr.io/stefanprodan/podinfo\' resolved to: 5.0.3","reason":"ReconciliationSucceeded","status":"True","type":"Ready"}],"latestImage":"ghcr.io/stefanprodan/podinfo:5.0.3","observedGeneration":2}}\n',
      clusterName: 'Default',
      tenant: '',
      uid: '5009e51d-0fee-4f8e-9df1-7684c8aac4bd',
      inventory: [],
    },
  ],
  errors: [],
};

const imageUpdateAutomations = {
  objects: [
    {
      payload:
        '{"apiVersion":"image.toolkit.fluxcd.io/v1beta1","kind":"ImageUpdateAutomation","metadata":{"creationTimestamp":"2023-02-06T15:26:21Z","finalizers":["finalizers.fluxcd.io"],"generation":1,"labels":{"kustomize.toolkit.fluxcd.io/name":"flux-system","kustomize.toolkit.fluxcd.io/namespace":"flux-system"},"managedFields":[{"apiVersion":"image.toolkit.fluxcd.io/v1beta1","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:labels":{"f:kustomize.toolkit.fluxcd.io/name":{},"f:kustomize.toolkit.fluxcd.io/namespace":{}}},"f:spec":{"f:git":{"f:checkout":{"f:ref":{"f:branch":{}}},"f:commit":{"f:author":{"f:email":{},"f:name":{}},"f:messageTemplate":{}},"f:push":{"f:branch":{}}},"f:interval":{},"f:sourceRef":{"f:kind":{},"f:name":{}},"f:update":{"f:path":{},"f:strategy":{}}}},"manager":"kustomize-controller","operation":"Apply","time":"2023-02-06T15:26:21Z"},{"apiVersion":"image.toolkit.fluxcd.io/v1beta1","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:finalizers":{".":{},"v:\\"finalizers.fluxcd.io\\"":{}}},"f:status":{"f:conditions":{},"f:observedGeneration":{}}},"manager":"image-automation-controller","operation":"Update","time":"2023-02-06T15:26:25Z"}],"name":"flux-system","namespace":"flux-system","resourceVersion":"1730794","uid":"84496ab1-2212-4b7f-8714-d90accd1633d"},"spec":{"git":{"checkout":{"ref":{"branch":"main"}},"commit":{"author":{"email":"fluxcdbot@users.noreply.github.com","name":"fluxcdbot"},"messageTemplate":"{{range .Updated.Images}}{{println .}}{{end}}"},"push":{"branch":"main"}},"interval":"1m0s","sourceRef":{"kind":"GitRepository","name":"flux-system"},"update":{"path":"./clusters/my-cluster","strategy":"Setters"}},"status":{"conditions":[{"lastTransitionTime":"2023-02-06T15:26:25Z","message":"unknown error: ERROR: The key you are authenticating with has been marked as read only.","reason":"ReconciliationFailed","status":"False","type":"Ready"}],"observedGeneration":1}}\n',
      clusterName: 'Default',
      tenant: '',
      uid: '84496ab1-2212-4b7f-8714-d90accd1633d',
      inventory: [],
    },
    {
      payload:
        '{"apiVersion":"image.toolkit.fluxcd.io/v1beta1","kind":"ImageUpdateAutomation","metadata":{"creationTimestamp":"2023-02-06T14:26:01Z","finalizers":["finalizers.fluxcd.io"],"generation":1,"labels":{"kustomize.toolkit.fluxcd.io/name":"flux-system","kustomize.toolkit.fluxcd.io/namespace":"flux-system"},"managedFields":[{"apiVersion":"image.toolkit.fluxcd.io/v1beta1","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:labels":{"f:kustomize.toolkit.fluxcd.io/name":{},"f:kustomize.toolkit.fluxcd.io/namespace":{}}},"f:spec":{"f:git":{"f:checkout":{"f:ref":{"f:branch":{}}},"f:commit":{"f:author":{"f:email":{},"f:name":{}},"f:messageTemplate":{}},"f:push":{"f:branch":{}}},"f:interval":{},"f:sourceRef":{"f:kind":{},"f:name":{},"f:namespace":{}},"f:update":{"f:path":{},"f:strategy":{}}}},"manager":"kustomize-controller","operation":"Apply","time":"2023-02-06T14:26:01Z"},{"apiVersion":"image.toolkit.fluxcd.io/v1beta1","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:finalizers":{".":{},"v:\\"finalizers.fluxcd.io\\"":{}}},"f:status":{"f:conditions":{},"f:lastAutomationRunTime":{},"f:observedGeneration":{}}},"manager":"image-automation-controller","operation":"Update","time":"2023-02-06T14:26:03Z"}],"name":"podinfo","namespace":"flux-system","resourceVersion":"1730796","uid":"ee3a5327-708a-49db-bcc1-07451bd92e18"},"spec":{"git":{"checkout":{"ref":{"branch":"main"}},"commit":{"author":{"email":"fluxcdbot@users.noreply.github.com","name":"fluxcdbot"},"messageTemplate":"{{range .Updated.Images}}{{println .}}{{end}}"},"push":{"branch":"main"}},"interval":"1m0s","sourceRef":{"kind":"GitRepository","name":"flux-system","namespace":"flux-system"},"update":{"path":"./clusters/my-cluster/deployment-podinfo.yaml","strategy":"Setters"}},"status":{"conditions":[{"lastTransitionTime":"2023-02-06T15:26:24Z","message":"unknown error: ERROR: The key you are authenticating with has been marked as read only.","reason":"ReconciliationFailed","status":"False","type":"Ready"}],"lastAutomationRunTime":"2023-02-06T15:26:16Z","observedGeneration":1}}\n',
      clusterName: 'Default',
      tenant: '',
      uid: 'ee3a5327-708a-49db-bcc1-07451bd92e18',
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
    core.ListObjectsReturns = {
      [Kind.ImageRepository]: imageRepositories,
      [Kind.ImagePolicy]: imagePolicies,
      [Kind.ImageUpdateAutomation]: imageUpdateAutomations,
    };
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

  it('renders image automation tabs and image-repo-list', async () => {
    core.IsCRDAvailableReturn = {
      'imageupdateautomations.image.toolkit.fluxcd.io': IsCRDAvailableReturn,
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
      ['Name', 'Namespace', 'Cluster', 'Status', 'Interval', 'Tag Count'],
      imageRepositories.objects.length,
    );
    const mappedObjects = imageRepositories.objects.map(
      obj => new ImageRepository(obj),
    );
    const search = mappedObjects[0].name;
    const searchedRows = mappedObjects
      .filter(e => e.name === search)
      .map((obj: any) => {
        return [
          obj.name,
          obj.namespace,
          obj.clusterName,
          obj?.obj?.status?.conditions[0].status === 'True'
            ? 'Ready'
            : 'Not Ready',
          showInterval(obj.interval),
          `${obj.tagCount}`,
        ];
      });
    filterTable.testSearchTableByValue(search, searchedRows);
    filterTable.clearSearchByVal(search);
  });

  it('renders Image Policies', async () => {
    await act(async () => {
      const c = wrap(<ImagePoliciesTable />);
      render(c);
    });
    const filterTable = new TestFilterableTable('image-policy-list', fireEvent);
    filterTable.testRenderTable(
      [
        'Name',
        'Namespace',
        'Cluster',
        'Status',
        'Image Policy',
        'Order/Range',
        'Image Repository',
      ],
      imagePolicies.objects.length,
    );
    const mappedObjects = imagePolicies.objects.map(
      obj => new ImagePolicy(obj),
    );
    const search = mappedObjects[0].name;
    const searchedRows = mappedObjects
      .filter(e => e.name === search)
      .map((obj: any) => {
        return [
          obj.name,
          obj.namespace,
          obj.clusterName,
          obj?.obj?.status?.conditions[0].status === 'True'
            ? 'Ready'
            : 'Not Ready',
          obj.imagePolicy?.type,
          obj.imagePolicy?.value,
          obj.imageRepositoryRef,
        ];
      });
    filterTable.testSearchTableByValue(search, searchedRows);
    filterTable.clearSearchByVal(search);
  });
  it('renders Image update automations', async () => {
    await act(async () => {
      const c = wrap(<ImageAutomationUpdatesTable />);
      render(c);
    });
    const filterTable = new TestFilterableTable('image-update-list', fireEvent);
    filterTable.testRenderTable(
      [
        'Name',
        'Namespace',
        'Cluster',
        'Status',
        'Source',
        'Interval',
        'Last Run',
      ],
      imageUpdateAutomations.objects.length,
    );
    const mappedObjects = imageUpdateAutomations.objects.map(
      obj => new ImageUpdateAutomation(obj),
    );
    const search = mappedObjects[1].name;
    const searchedRows = mappedObjects
      .filter(e => e.name === search)
      .map((obj: any) => {
        return [
          obj.name,
          obj.namespace,
          obj.clusterName,
          obj?.obj?.status?.conditions[0].status === 'True'
            ? 'Ready'
            : 'Not Ready',
          `${obj.sourceRef.kind}/${obj.sourceRef.name}`,
          showInterval(obj.interval),
          obj.lastAutomationRunTime
            ? moment(obj.lastAutomationRunTime).fromNow()
            : '',
        ];
      });
    filterTable.testSearchTableByValue(search, searchedRows);
    filterTable.clearSearchByVal(search);
  });
});
