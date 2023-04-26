import { ClustersService } from '../../../cluster-services/cluster_services.pb';
import { QueryClient, useQueryClient } from 'react-query';

interface DetailParams {
  clusterName: string;
  namespace: string;
  externalSecretName: string;
}

const SECRETS_KEY = 'external-secrets';

function invalidate(
  qc: QueryClient,
  { clusterName, namespace, externalSecretName }: DetailParams,
) {
  return qc.invalidateQueries([
    SECRETS_KEY,
    clusterName,
    namespace,
    externalSecretName,
  ]);
}

export function useSyncSecret(params: DetailParams) {
  const qc = useQueryClient();

  return () =>
    ClustersService.SyncExternalSecrets(params).then(res => {
      invalidate(qc, params);
      return res;
    });
}
