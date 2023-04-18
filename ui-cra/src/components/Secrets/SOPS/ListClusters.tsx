import { MenuItem } from '@material-ui/core';
import React from 'react';
import { useListCluster } from '../../../hooks/clusters';
import { Select } from '../../../utils/form';
import LoadingWrapper from '../../Workspaces/WorkspaceDetails/Tabs/WorkspaceTabsWrapper';

const ListClusters = ({
  value,
  validateForm,
  handleFormData,
}: {
  value: string;
  validateForm: boolean;
  handleFormData: (value: any) => void;
}) => {
  const { isLoading, data, error } = useListCluster();
  return (
    <LoadingWrapper loading={isLoading} errorMessage={error?.message}>
      <Select
        className="form-section"
        name="clusterName"
        required={true}
        label="CLUSTER"
        onChange={event => handleFormData(event.target.value)}
        value={value}
        error={validateForm && !value}
      >
        {data?.gitopsClusters
          ?.filter(e =>
            e.conditions?.find(c => c.status === 'True' && c.type === 'Ready'),
          )
          .map((option, index: number) => {
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
