import { MenuItem } from '@material-ui/core';
import {
    RequestStateHandler,
    Text,
    useListAutomations,
} from '@weaveworks/weave-gitops';
import { RequestError } from '../../../types/custom';
import { Select } from '../../../utils/form';

const ListApplications = ({
  value,
  validateForm,
  handleFormData,
}: {
  value: string;
  validateForm: boolean;
  handleFormData: (value: any) => void;
}) => {
  const { data, isLoading, error } = useListAutomations('', { retry: false });
  return (
    <RequestStateHandler loading={isLoading} error={error as RequestError}>
      <Select
        name="applicationName"
        required={true}
        label="SELECT THE APPLICATION YOU WANT TO PROMOTE"
        onChange={event => handleFormData(event.target.value)}
        value={value}
        error={validateForm && !value}
      >
        {data?.result?.length ? (
          data?.result
            ?.filter(e =>
              e.conditions?.find(
                c => c.status === 'True' && c.type === 'Ready',
              ),
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
                  <Text>{option.name}</Text>
                </MenuItem>
              );
            })
        ) : (
          <MenuItem value="" disabled={true}>
            No Applications found in
          </MenuItem>
        )}
      </Select>
    </RequestStateHandler>
  );
};

export default ListApplications;
