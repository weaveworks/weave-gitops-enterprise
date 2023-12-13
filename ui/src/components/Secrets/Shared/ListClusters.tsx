import { MenuItem } from '@material-ui/core';
import { Flex, RequestStateHandler, Text } from '@weaveworks/weave-gitops';
import { RequestError } from '@weaveworks/weave-gitops/ui/lib/types';
import { useListCluster } from '../../../hooks/clusters';
import { Select } from '../../../utils/form';

const ListClusters = ({
  value,
  error,
  handleFormData,
}: {
  value: string;
  error: boolean;
  handleFormData: (value: any) => void;
}) => {
  const { isLoading, data, error: listError } = useListCluster();
  return (
    <RequestStateHandler loading={isLoading} error={listError as RequestError}>
      <Select
        name="clusterName"
        required={true}
        label="CLUSTER"
        onChange={event => handleFormData(event.target.value)}
        value={value}
        error={error}
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
                <Flex column>
                  <Text>{option.name}</Text>
                  <Text color="neutral30" size="small">
                    {option.namespace ? `ns: ${option.namespace}` : '-'}
                  </Text>
                </Flex>
              </MenuItem>
            );
          })}
      </Select>
    </RequestStateHandler>
  );
};

export default ListClusters;
