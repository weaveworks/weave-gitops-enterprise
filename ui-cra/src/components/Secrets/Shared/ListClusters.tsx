import { MenuItem } from '@material-ui/core';
import { RequestStateHandler } from '@weaveworks/weave-gitops';
import { RequestError } from '@weaveworks/weave-gitops/ui/lib/types';
import { useListCluster } from '../../../hooks/clusters';
import { Select } from '../../../utils/form';

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
    <RequestStateHandler loading={isLoading} error={error as RequestError}>
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
    </RequestStateHandler>
  );
};

export default ListClusters;
