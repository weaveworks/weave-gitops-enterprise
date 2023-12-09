import { CoreClientContext, FluxObject } from '@weaveworks/weave-gitops';
import { ListError } from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';
import { RequestError } from '@weaveworks/weave-gitops/ui/lib/types';
import { useContext } from 'react';
import { QueryClient, useQuery, useQueryClient } from 'react-query';
import useNotifications from '../contexts/Notifications';

const GITOPSSETS_KEY = 'gitopssets';
const GITOPSSETS_POLL_INTERVAL = 5000;

type Res = { objects: FluxObject[]; errors: ListError[] };

export function useListGitOpsSets(
  opts: { enabled: boolean } = {
    enabled: true,
  },
) {
  const { setNotifications } = useNotifications();
  const { api } = useContext(CoreClientContext);

  const onError = (error: Error) =>
    setNotifications([{ message: { text: error.message }, severity: 'error' }]);

  return useQuery<Res, Error>(
    [GITOPSSETS_KEY],
    async () => {
      const res = await api.ListObjects({ kind: 'GitOpsSet' });
      let objects: FluxObject[] = [];
      if (res.objects) {
        objects = res.objects.map(obj => new FluxObject(obj));
      }
      return { objects, errors: res.errors || [] };
    },
    {
      keepPreviousData: true,
      refetchInterval: GITOPSSETS_POLL_INTERVAL,
      onError,
      ...opts,
    },
  );
}

interface DetailParams {
  name: string;
  namespace: string;
  clusterName: string;
}

export function useGetGitOpsSet({
  name,
  namespace,
  clusterName,
}: DetailParams) {
  const { setNotifications } = useNotifications();
  const { api } = useContext(CoreClientContext);
  const onError = (error: Error) =>
    setNotifications([{ message: { text: error.message }, severity: 'error' }]);

  return useQuery<FluxObject, RequestError>(
    [GITOPSSETS_KEY, clusterName, namespace, name],
    async () => {
      const res = await api.GetObject({
        name,
        namespace,
        clusterName,
        kind: 'GitOpsSet',
      });
      return new FluxObject(res.object!);
    },
    {
      onError,
    },
  );
}

interface DetailParams {
  name: string;
  namespace: string;
  clusterName: string;
}

function invalidate(
  qc: QueryClient,
  { name, namespace, clusterName }: DetailParams,
) {
  return qc.invalidateQueries([GITOPSSETS_KEY, clusterName, namespace, name]);
}

export function useSyncGitOpsSet(params: DetailParams) {
  const qc = useQueryClient();
  const { api } = useContext(CoreClientContext);

  return () =>
    api
      .SyncFluxObject({ objects: [{ kind: 'GitOpsSet', ...params }] })
      .then(res => {
        invalidate(qc, params);
        return res;
      });
}

export function useToggleSuspendGitOpsSet(params: DetailParams) {
  const qc = useQueryClient();
  const { api } = useContext(CoreClientContext);

  return (suspend: boolean) =>
    api
      .ToggleSuspendResource({
        objects: [{ kind: 'GitOpsSet', ...params }],
        suspend,
      })
      .then(res => {
        return invalidate(qc, params).then(() => res);
      });
}
