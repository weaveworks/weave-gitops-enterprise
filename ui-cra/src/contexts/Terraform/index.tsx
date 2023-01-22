import { ListError } from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';
import { RequestError } from '@weaveworks/weave-gitops/ui/lib/types';
import * as React from 'react';
import { QueryClient, useQuery, useQueryClient } from 'react-query';
import {
  GetTerraformObjectResponse,
  GetTerraformObjectPlanResponse,
  ListTerraformObjectsResponse,
  Terraform,
} from '../../api/terraform/terraform.pb';
import { formatError } from '../../utils/formatters';
import useNotifications from './../../contexts/Notifications';

const TerraformContext = React.createContext<typeof Terraform>(
  {} as typeof Terraform,
);

interface Props {
  api: typeof Terraform;
  children?: any;
}

export function TerraformProvider({ api, children }: Props) {
  return (
    <TerraformContext.Provider value={api}>
      {children}
    </TerraformContext.Provider>
  );
}

function useTerraform() {
  return React.useContext(TerraformContext);
}

const TERRAFORM_KEY = 'terraform';
const TERRAFORM_PLAN_KEY = 'terraform_plan';

export function useListTerraformObjects() {
  const tf = useTerraform();

  return useQuery<ListTerraformObjectsResponse, ListError>(
    [TERRAFORM_KEY],
    () => tf.ListTerraformObjects({}),
    {
      retry: false,
      refetchInterval: 5000,
    },
  );
}

interface DetailParams {
  name: string;
  namespace: string;
  clusterName: string;
}

export function useGetTerraformObjectDetail(
  { name, namespace, clusterName }: DetailParams,
  enabled?: boolean,
) {
  const tf = useTerraform();
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));

  return useQuery<GetTerraformObjectResponse, RequestError>(
    [TERRAFORM_KEY, clusterName, namespace, name],
    () => tf.GetTerraformObject({ name, namespace, clusterName }),
    { onError, enabled, refetchInterval: 5000 },
  );
}

export function useGetTerraformObjectPlan(
  { name, namespace, clusterName }: DetailParams,
  enabled?: boolean,
) {
  const tf = useTerraform();

  const onError = (error: Error) => {};

  return useQuery<GetTerraformObjectPlanResponse, RequestError>(
    [TERRAFORM_PLAN_KEY, clusterName, namespace, name],
    () => tf.GetTerraformObjectPlan({ name, namespace, clusterName }),
    { onError, enabled, refetchInterval: 5000 },
  );
}

function invalidate(
  qc: QueryClient,
  { name, namespace, clusterName }: DetailParams,
) {
  return qc.invalidateQueries([TERRAFORM_KEY, clusterName, namespace, name]);
}

export function useSyncTerraformObject(params: DetailParams) {
  const tf = useTerraform();
  const qc = useQueryClient();

  return () =>
    tf.SyncTerraformObject(params).then(res => {
      invalidate(qc, params);

      return res;
    });
}

export function useToggleSuspendTerraformObject(params: DetailParams) {
  const tf = useTerraform();
  const qc = useQueryClient();

  return (suspend: boolean) =>
    tf.ToggleSuspendTerraformObject({ ...params, suspend }).then(res => {
      return invalidate(qc, params).then(() => res);
    });
}
