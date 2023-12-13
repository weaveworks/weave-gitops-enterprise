import { MenuItem } from '@material-ui/core';
import { RequestStateHandler } from '@weaveworks/weave-gitops';
import { RequestError } from '@weaveworks/weave-gitops/ui/lib/types';
import { useListKustomizationSOPS } from '../../../hooks/listSOPSKustomization';
import { Select } from '../../../utils/form';

const ListKustomizations = ({
  value,
  error,
  handleFormData,
  clusterName,
}: {
  value: string;
  error: boolean;
  handleFormData: (value: any) => void;
  clusterName: string;
}) => {
  const {
    isLoading,
    error: listError,
    data,
  } = useListKustomizationSOPS(
    { clusterName },
    {
      retry: false,
    },
  );

  return (
    <RequestStateHandler loading={isLoading} error={listError as RequestError}>
      <Select
        className="form-section"
        required
        name="kustomization"
        label="KUSTOMIZATION"
        description="Choose the kustomization that will be used by SOPS to decrypt the secret."
        onChange={event => handleFormData(event.target.value)}
        value={value}
        error={error}
      >
        {data?.kustomizations?.length ? (
          data?.kustomizations?.map((k, index: number) => {
            return (
              <MenuItem key={index} value={`${k.name}/${k.namespace}`}>
                {k.name}
              </MenuItem>
            );
          })
        ) : (
          <MenuItem value="" disabled={true}>
            No Kustomization found in {clusterName}
          </MenuItem>
        )}
      </Select>
    </RequestStateHandler>
  );
};

export default ListKustomizations;
