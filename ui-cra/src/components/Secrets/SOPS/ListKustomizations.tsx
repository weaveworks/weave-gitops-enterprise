import { MenuItem } from '@material-ui/core';
import { useListKustomizationSOPS } from '../../../hooks/listSOPSKustomization';
import { Select } from '../../../utils/form';
import LoadingWrapper from '../../Workspaces/WorkspaceDetails/Tabs/WorkspaceTabsWrapper';

const ListKustomizations = ({
  value,
  handleFormData,
  clusterName,
}: {
  value: string;
  handleFormData: (value: any) => void;
  clusterName: string;
}) => {
  const { isLoading, error, data } = useListKustomizationSOPS(
    { clusterName },
    {
      retry: false,
    },
  );

  return (
    <LoadingWrapper loading={isLoading} errorMessage={error?.message}>
      <Select
        className="form-section"
        required
        name="kustomization"
        label="KUSTOMIZATION"
        description="Choose the kustomization that will be used by SOPS to decrypt the secret."
        onChange={event => handleFormData(event.target.value)}
        value={value}
      >
        {!!data?.kustomizations?.length ? (
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
    </LoadingWrapper>
  );
};

export default ListKustomizations;
