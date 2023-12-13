import { MenuItem } from '@material-ui/core';
import { RequestStateHandler } from '@weaveworks/weave-gitops';
import { useListExternalSecretStores } from '../../../contexts/Secrets';
import { RequestError } from '../../../types/custom';
import { Select } from '../../../utils/form';

const ListSecretsStore = ({
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
    data,
    isLoading,
    error: listError,
  } = useListExternalSecretStores({
    clusterName,
  });
  return (
    <RequestStateHandler loading={isLoading} error={listError as RequestError}>
      <Select
        required
        name="secretStoreRef"
        label="SECRET STORE"
        onChange={event => handleFormData(event.target.value)}
        value={value}
        error={error}
      >
        {data?.stores?.length ? (
          data?.stores?.map((s, index: number) => {
            return (
              <MenuItem
                key={index}
                value={`${s.name}/${s.kind}/${s.namespace}/${s.type}`}
              >
                {s.name}
              </MenuItem>
            );
          })
        ) : (
          <MenuItem value="" disabled={true}>
            No SecretStore found in {clusterName}
          </MenuItem>
        )}
      </Select>
    </RequestStateHandler>
  );
};

export default ListSecretsStore;
