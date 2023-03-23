import { MenuItem } from '@material-ui/core';
import { useListKustomizationSOPS } from '../../../hooks/listObjects';
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
        description={
          !!data?.kustomizations?.length
            ? 'Choose the kustomization that will be used by SOPS to decrypt the secret.'
            : `No Kustomization found in ${clusterName}`
        }
        onChange={event => handleFormData(event.target.value)}
        value={value}
        disabled={!clusterName || !data?.kustomizations?.length}
      >
        {data?.kustomizations?.map((k, index: number) => {
          return (
            <MenuItem key={index} value={`${k.name}/${k.namespace}`}>
              {k.name}
            </MenuItem>
          );
        })}
      </Select>
    </LoadingWrapper>
  );
};

export default ListKustomizations;
