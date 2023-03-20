import { MenuItem } from '@material-ui/core';
import React from 'react';
import { useListCluster } from '../../../hooks/clusters';
import { Select } from '../../../utils/form';
import LoadingWrapper from '../../Workspaces/WorkspaceDetails/Tabs/WorkspaceTabsWrapper';

const ListClusters = ({
  value,
  handleFormData,
}: {
  value: string;
  handleFormData: (value: any) => void;
}) => {
  let { isLoading, data, error } = useListCluster();
  return (
    <LoadingWrapper loading={isLoading} errorMessage={error?.message}>
      <Select
        className="form-section"
        name="clusterName"
        required
        label="CLUSTER"
        onChange={event => handleFormData(event.target.value)}
        value={value}
      >
        {data?.gitopsClusters?.map((option, index: number) => {
          return (
            <MenuItem
              key={index}
              value={
                option.namespace
                  ? `${option.namespace}/${option.name}`
                  : option.name
              }
            >
              {option.name}
            </MenuItem>
          );
        })}
      </Select>
    </LoadingWrapper>
  );
};

export default ListClusters;
