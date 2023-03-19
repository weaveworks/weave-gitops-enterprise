import React from 'react';
import { MenuItem } from '@material-ui/core';
import { Kind, Kustomization } from '@weaveworks/weave-gitops';
import { useListObjects } from '../../../hooks/listObjects';
import { Select } from '../../../utils/form';
import LoadingWrapper from '../../Workspaces/WorkspaceDetails/Tabs/WorkspaceTabsWrapper';

const ListKustomizations = ({
  value,
  handleFormData,
}: {
  value: string;
  handleFormData: (value: any) => void;
}) => {
  const {
    isLoading,
    error,
    data: kustomizations,
  } = useListObjects(Kustomization, Kind.Kustomization, '', { retry: false });
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
        {kustomizations?.objects?.map((k, index: number) => {
          return (
            <MenuItem key={index} value={k.name}>
              {k.name}
            </MenuItem>
          );
        })}
      </Select>
    </LoadingWrapper>
  );
};

export default ListKustomizations;
