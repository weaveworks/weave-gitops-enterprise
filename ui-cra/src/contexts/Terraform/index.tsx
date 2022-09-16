import { ListError } from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';
import { RequestError } from '@weaveworks/weave-gitops/ui/lib/types';
import * as React from 'react';
import { useQuery } from 'react-query';
import {
  GetTerraformObjectResponse,
  ListTerraformObjectsResponse,
  Terraform,
} from '../../api/terraform/terraform.pb';

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

export function useListTerraformObjects() {
  const tf = useTerraform();

  return useQuery<ListTerraformObjectsResponse, ListError>(
    [TERRAFORM_KEY],
    () => tf.ListTerraformObjects({}),
    {
      // fetch once only
      retry: false,
      cacheTime: Infinity,
      staleTime: Infinity,
    },
  );
}

export function useTerraformObjectCount() {
  const { data } = useListTerraformObjects();

  return data?.objects?.length || 0;
}

interface DetailParams {
  name: string;
  namespace: string;
  clusterName: string;
}

export function useGetTerraformObjectDetail({
  name,
  namespace,
  clusterName,
}: DetailParams) {
  const tf = useTerraform();

  return useQuery<GetTerraformObjectResponse, RequestError>(
    [TERRAFORM_KEY, clusterName, namespace, name],
    () => tf.GetTerraformObject({ name, namespace, clusterName }),
  );
}
