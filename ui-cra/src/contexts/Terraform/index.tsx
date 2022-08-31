import * as React from 'react';
import { useQuery } from 'react-query';
import { Terraform } from '../../api/terraform/terraform.pb';

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

  return useQuery([TERRAFORM_KEY], () => tf.ListTerraformObjects({}), {
    // fetch once only
    retry: false,
    cacheTime: Infinity,
    staleTime: Infinity,
  });
}
