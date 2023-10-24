import { act, render, screen } from '@testing-library/react';
import EnterpriseClientProvider from '../../../../contexts/EnterpriseClient/Provider';
import {
  ClustersServiceClientMock,
  defaultContexts,
  withContext,
} from '../../../../utils/test-utils';
import PreviewModal from '../Partials/PreviewModal';

Object.assign(navigator, {
  clipboard: {
    writeText: () => {
      return;
    },
  },
});

describe('PR Preview when creating resources', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: ClustersServiceClientMock;
  beforeEach(() => {
    let clipboardContents = '';

    Object.assign(navigator, {
      clipboard: {
        writeText: (text: string) => {
          clipboardContents = text;
          return Promise.resolve(text);
        },
        readText: () => Promise.resolve(clipboardContents),
      },
    });

    api = new ClustersServiceClientMock();
    wrap = withContext([
      ...defaultContexts(),
      [EnterpriseClientProvider, { api }],
    ]);
  });

  it('renders the PR Preview Modal', async () => {
    const prPreview = {
      renderedTemplates: [
        {
          path: 'clusters/management/clusters/default/test.yaml',
          content:
            'apiVersion: gitops.weave.works/v1alpha1\nkind: GitopsCluster\nmetadata:\n  labels:\n    templates.weave.works/template-name: vcluster-template-development\n    templates.weave.works/template-namespace: default\n    weave.works/capi: bootstrap\n  name: test\n  namespace: default\n  annotations:\n    templates.weave.works/created-files: "{\\"files\\":[\\"clusters/management/clusters/default/test.yaml\\"]}"\nspec:\n  capiClusterRef:\n    name: test\n\n---\napiVersion: cluster.x-k8s.io/v1beta1\nkind: Cluster\nmetadata:\n  labels:\n    templates.weave.works/template-name: vcluster-template-development\n    templates.weave.works/template-namespace: default\n  name: test\n  namespace: default\nspec:\n  controlPlaneRef:\n    apiVersion: infrastructure.cluster.x-k8s.io/v1alpha1\n    kind: VCluster\n    name: test\n  infrastructureRef:\n    apiVersion: infrastructure.cluster.x-k8s.io/v1alpha1\n    kind: VCluster\n    name: test\n\n---\napiVersion: infrastructure.cluster.x-k8s.io/v1alpha1\nkind: VCluster\nmetadata:\n  labels:\n    templates.weave.works/template-name: vcluster-template-development\n    templates.weave.works/template-namespace: default\n  name: test\n  namespace: default\nspec:\n  helmRelease:\n    values: |\n      syncer:\n        extraArgs:\n          - "--tls-san=test.default.svc"\n  kubernetesVersion: 1.23.3\n',
        },
      ],
      profileFiles: [],
      kustomizationFiles: [
        {
          path: 'clusters/default/test/clusters-bases-kustomization.yaml',
          content:
            'apiVersion: kustomize.toolkit.fluxcd.io/v1\nkind: Kustomization\nmetadata:\n  creationTimestamp: null\n  name: clusters-bases-kustomization\n  namespace: flux-system\nspec:\n  interval: 10m0s\n  path: clusters/bases\n  prune: true\n  sourceRef:\n    kind: GitRepository\n    name: flux-system\nstatus: {}\n',
        },
      ],
      externalSecretsFiles: [],
      policyConfigFiles: [],
      sopsSecretFiles: [],
    };

    api.RenderTemplateReturns = prPreview;

    await act(async () => {
      const c = wrap(
        <PreviewModal
          openPreview={true}
          setOpenPreview={() => console.log('')}
          prPreview={prPreview}
        />,
      );
      render(c);
    });
    expect(await screen.findByText('PR Preview')).toBeTruthy();

    // enabled tabs

    const tab1Title = screen.getByRole('tab', {
      name: /Resource Definition/i,
    });
    expect(tab1Title).toBeInTheDocument();

    const tab1Content = screen.getByTestId('tab-content-0');
    expect(tab1Content.textContent).toEqual(
      api?.RenderTemplateReturns?.renderedTemplates?.[0].content,
    );

    const tab3Title = screen.getByRole('tab', {
      name: /Kustomizations/i,
    });
    expect(tab3Title).toBeInTheDocument();

    // navigate to the Kustomizations tab
    tab3Title.click();

    const tab3Content = screen.getByTestId('tab-content-2');
    expect(tab3Content.textContent).toEqual(
      api?.RenderTemplateReturns?.kustomizationFiles?.[0].content,
    );

    // disabled tabs

    expect(screen.getByText(/Profiles/i).closest('button')).toHaveAttribute(
      'disabled',
    );
  });
});
