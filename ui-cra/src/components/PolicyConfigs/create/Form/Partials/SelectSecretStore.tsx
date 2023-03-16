import { MenuItem } from '@material-ui/core';
import { Dispatch, useEffect } from 'react';
import { ExternalSecretStore } from '../../../../../cluster-services/cluster_services.pb';
import { Input, Select } from '../../../../../utils/form';
import { useGetSecretStoreDetails } from '../../../../../contexts/Secrets';

interface SelectSecretStoreProps {
  cluster: string;
  handleFormData: (
    event: React.ChangeEvent<{ name?: string; value: unknown }>,
    fieldName: string,
  ) => void;
  formError: string;
  selectedSecretStore: ExternalSecretStore;
  setSelectedSecretStore: Dispatch<React.SetStateAction<any>>;
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
  automation: any;
}

export const SelectSecretStore = (props: SelectSecretStoreProps) => {
  const {
    cluster,
    handleFormData,
    formError,
    selectedSecretStore,
    setSelectedSecretStore,
    formData,
    setFormData,
    automation,
  } = props;
  const { data, isLoading } = useGetSecretStoreDetails({
    clusterName: cluster,
  });
  const { secretStoreRef, secretNamespace, secretStoreType } = automation;
  const secretStores = data?.stores;

  useEffect(() => {
    if (secretStoreRef) {
      const selectedStore =
        secretStores?.find(item => item.name === secretStoreRef) || {};
      setSelectedSecretStore(selectedStore);
    }
  },[secretStores, secretStoreRef, setSelectedSecretStore]);

  const handleSelectSecretStore = (event: React.ChangeEvent<any>) => {
    const sercetStore = event.target.value;
    const value = JSON.parse(sercetStore);
    setSelectedSecretStore(value);

    let currentAutomation = [...formData.clusterAutomations];
    currentAutomation[0] = {
      ...automation,
      secretStoreRef: value.name,
      secretNamespace: value.namespace,
      secretStoreType: value.type,
      secretStoreKind: value.kind,
    };

    setFormData({
      ...formData,
      clusterAutomations: currentAutomation,
    });
  };
  
  return (
    <div className="form-group">
      <Select
        className="form-section"
        name="secretStoreRef"
        required
        label="SECRET STORE"
        value={
          !!Object.keys(selectedSecretStore).length
            ? JSON.stringify(selectedSecretStore)
            : ''
        }
        onChange={handleSelectSecretStore}
        error={formError === 'secretStoreRef' && !secretStoreRef}
      >
        {isLoading ? (
          <MenuItem disabled={true}>Loading...</MenuItem>
        ) : secretStores?.length ? (
          secretStores.map((option, index: number) => {
            return (
              <MenuItem key={index} value={JSON.stringify(option)}>
                {option.name}
              </MenuItem>
            );
          })
        ) : (
          <MenuItem disabled={true}>
            No SecretStore found on that cluster
          </MenuItem>
        )}
      </Select>
      <Input
        className="form-section"
        name="secret_store_kind"
        label="SECRET STORE TYPE"
        value={secretStoreType}
        disabled={true}
        error={false}
      />
      <Input
        className="form-section"
        required
        name="secretNamespace"
        label="TARGET NAMESPACE"
        value={secretNamespace}
        disabled={!!selectedSecretStore?.namespace ? true : false}
        onChange={event => handleFormData(event, 'secretNamespace')}
        error={formError === 'secretNamespace' && !secretNamespace}
      />
    </div>
  );
};
