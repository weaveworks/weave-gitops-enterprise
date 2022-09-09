import {
  FluxObjectKind,
  V2Routes,
} from '@weaveworks/weave-gitops';

export function getKindRoute(k: FluxObjectKind | string): string {
  switch (k) {
    case FluxObjectKind.KindGitRepository:
    case 'GitRepository':
      return V2Routes.GitRepo;

    case FluxObjectKind.KindBucket:
    case 'Bucket':
      return V2Routes.Bucket;

    case FluxObjectKind.KindHelmRepository:
    case 'HelmRepository':
      return V2Routes.HelmRepo;

    case FluxObjectKind.KindHelmChart:
    case 'HelmChart':
      return V2Routes.HelmChart;

    case FluxObjectKind.KindKustomization:
    case 'Kustomization':
      return V2Routes.Kustomization;

    case FluxObjectKind.KindHelmRelease:
    case 'HelmRelease':
      return V2Routes.HelmRelease;

    default:
      return V2Routes.NotImplemented;
  }
}
