import {
  Bucket,
  FluxObject,
  GitRepository,
  HelmChart,
  HelmRepository,
  Kind,
  OCIRepository,
  PARENT_CHILD_LOOKUP,
  Alert,
  HelmRelease,
  Kustomization,
  Provider,
  coreClient,
} from '@weaveworks/weave-gitops';
import { RequestError } from '@weaveworks/weave-gitops/ui/lib/types';
import _ from 'lodash';
import { QueryClient, useQuery, useQueryClient } from 'react-query';
import {
  GetGitOpsSetResponse,
  GitOpsSets,
  ListGitOpsSetsResponse,
} from '../api/gitopssets/gitopssets.pb';
import {
  GroupVersionKind,
  Object as ResponseObject,
} from '../api/gitopssets/types.pb';
import useNotifications from '../contexts/Notifications';

const GITOPSSETS_KEY = 'gitopssets';
const GITOPSSETS_POLL_INTERVAL = 5000;

export function useListGitOpsSets(
  opts: { enabled: boolean } = {
    enabled: true,
  },
) {
  const { setNotifications } = useNotifications();

  const onError = (error: Error) =>
    setNotifications([{ message: { text: error.message }, severity: 'error' }]);

  return useQuery<ListGitOpsSetsResponse, Error>(
    [GITOPSSETS_KEY],
    () => GitOpsSets.ListGitOpsSets({}),
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
  const onError = (error: Error) =>
    setNotifications([{ message: { text: error.message }, severity: 'error' }]);

  return useQuery<GetGitOpsSetResponse, RequestError>(
    [GITOPSSETS_KEY, clusterName, namespace, name],
    () => GitOpsSets.GetGitOpsSet({ name, namespace, clusterName }),
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

  return () =>
    GitOpsSets.SyncGitOpsSet(params).then(res => {
      invalidate(qc, params);
      return res;
    });
}

export function useToggleSuspendGitOpsSet(params: DetailParams) {
  const qc = useQueryClient();

  return (suspend: boolean) =>
    GitOpsSets.ToggleSuspendGitOpsSet({ ...params, suspend }).then(res => {
      return invalidate(qc, params).then(() => res);
    });
}

export function useGetReconciledTree(
  name: string,
  namespace: string,
  type: 'GitOpsSet',
  kinds: GroupVersionKind[],
  clusterName = 'Default',
) {
  return useQuery<any[], RequestError>(
    ['reconciled_objects', { name, namespace, type, kinds }],
    () =>
      getChildren(
        coreClient,
        GitOpsSets,
        name,
        namespace,
        type,
        kinds,
        clusterName,
      ),
    { retry: false, refetchInterval: 5000 },
  );
}

export const getChildren = async (
  core: typeof coreClient,
  client: typeof GitOpsSets,
  name: string,
  namespace: string,
  automationKind: string,
  kinds: GroupVersionKind[],
  clusterName: string,
): Promise<FluxObject[]> => {
  const { objects } = await client.GetReconciledObjects({
    name,
    namespace,
    automationKind,
    kinds,
    clusterName,
  });
  const length = objects?.length || 0;
  const result = [];
  for (let o = 0; o < length; o++) {
    const obj = convertResponse('', objects?.[o] || ({} as ResponseObject));
    await getChildrenRecursive(
      core,
      namespace,
      obj,
      clusterName,
      PARENT_CHILD_LOOKUP,
    );
    result.push(obj);
  }
  return _.flatten(result);
};

export const getChildrenRecursive = async (
  core: typeof coreClient,
  namespace: string,
  object: FluxObject,
  clusterName: string,
  lookup: any,
) => {
  const children = [];

  const k = lookup[object?.type || ''];

  if (k && k.children) {
    for (let i = 0; i < k.children.length; i++) {
      const child: GroupVersionKind = k.children[i];

      const res = await core.GetChildObjects({
        parentUid: object.uid,
        namespace,
        groupVersionKind: child,
        clusterName: clusterName,
      });

      const length = res?.objects?.length || 0;

      for (let q = 0; q < length; q++) {
        const c = convertResponse(
          '',
          res?.objects?.[q] || ({} as ResponseObject),
        );
        // Dive down one level and update the lookup accordingly.
        await getChildrenRecursive(core, namespace, c, clusterName, {
          [child.kind as string]: child,
        });
        children.push(c);
      }
    }
  }
  object.children = children;
};

export function convertResponse(kind: Kind | string, response: ResponseObject) {
  if (kind === Kind.HelmRepository) {
    return new HelmRepository(response);
  }
  if (kind === Kind.HelmChart) {
    return new HelmChart(response);
  }
  if (kind === Kind.Bucket) {
    return new Bucket(response);
  }
  if (kind === Kind.GitRepository) {
    return new GitRepository(response);
  }
  if (kind === Kind.OCIRepository) {
    return new OCIRepository(response);
  }
  if (kind === Kind.Kustomization) {
    return new Kustomization(response);
  }
  if (kind === Kind.HelmRelease) {
    return new HelmRelease(response);
  }
  if (kind === Kind.Provider) {
    return new Provider(response);
  }
  if (kind === Kind.Alert) {
    return new Alert(response);
  }

  return new FluxObject(response);
}
