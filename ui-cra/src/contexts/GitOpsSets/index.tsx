import { RequestError } from '@weaveworks/weave-gitops/ui/lib/types';
import useNotifications from '../../contexts/Notifications';
import * as React from 'react';
import { QueryClient, useQuery, useQueryClient } from 'react-query';
import {
  GitOpsSets,
  ListGitOpsSetsResponse,
} from '../../api/gitopssets/gitopssets.pb';
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
} from '@weaveworks/weave-gitops';
import {
  GroupVersionKind,
  Object as ResponseObject,
  ResourceRef,
} from '../../api/gitopssets/types.pb';
import _ from 'lodash';

const GitOpsSetsContext = React.createContext<typeof GitOpsSets>(
  {} as typeof GitOpsSets,
);

interface Props {
  api: typeof GitOpsSets;
  children?: any;
}

export function GitOpsSetsProvider({ api, children }: Props) {
  return (
    <GitOpsSetsContext.Provider value={api}>
      {children}
    </GitOpsSetsContext.Provider>
  );
}

function useGitOpsSets() {
  return React.useContext(GitOpsSetsContext);
}

const GITOPSSETS_KEY = 'gitopssets';

export function useListGitOpsSets() {
  const { setNotifications } = useNotifications();

  const onError = (error: Error) =>
    setNotifications([{ message: { text: error.message }, severity: 'error' }]);

  return useQuery<ListGitOpsSetsResponse, Error>(
    [GITOPSSETS_KEY],
    () => GitOpsSets.ListGitOpsSets({}),
    {
      keepPreviousData: true,
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

// export function useSyncGitOpsSet(params: DetailParams) {
//   const gs = useGitOpsSets();
//   const qc = useQueryClient();

//   return () =>
//     gs.SyncGitOpsSet(params).then(res => {
//       invalidate(qc, params);

//       return res;
//     });
// }

export function useToggleSuspendGitOpsSet(params: DetailParams) {
  const gs = useGitOpsSets();
  const qc = useQueryClient();

  return (suspend: boolean) =>
    gs.ToggleSuspendGitOpsSet({ ...params, suspend }).then(res => {
      return invalidate(qc, params).then(() => res);
    });
}

export function useGetReconciledTree(
  name: string,
  namespace: string,
  type: 'GitOpsSet',
  kinds: ResourceRef[],
  clusterName = 'Default',
) {
  return useQuery<any[], RequestError>(
    ['reconciled_objects', { name, namespace, type, kinds }],
    () => getChildren(GitOpsSets, name, namespace, type, kinds, clusterName),
    { retry: false, refetchInterval: 5000 },
  );
}

export const getChildren = async (
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
      client,
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
  client: typeof GitOpsSets,
  namespace: string,

  object: FluxObject,
  clusterName: string,
  lookup: any,
) => {
  const children = [];

  const k = lookup[object.type || 'GitOpsSet'];

  if (k && k.children) {
    for (let i = 0; i < k.children.length; i++) {
      const child: GroupVersionKind = k.children[i];

      const res = await client.GetChildObjects({
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
        await getChildrenRecursive(client, namespace, c, clusterName, {
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
