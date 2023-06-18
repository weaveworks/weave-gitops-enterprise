import { Switch } from '@material-ui/core';
import RemoveCircleOutlineIcon from '@material-ui/icons/RemoveCircleOutline';
import { Button, Flex, Icon, IconType, Text } from '@weaveworks/weave-gitops';
import { Dispatch } from 'react';
import { InputDebounced } from '../../../utils/form';
import { ExternalSecret } from '../Shared/utils';

export const SecretProperty = ({
  formData,
  validateForm,
  setFormData,
}: {
  formData: ExternalSecret;
  validateForm: boolean;
  setFormData: Dispatch<React.SetStateAction<any>>;
}) => {
  const handleSecretChange = (id: number, isKey: boolean, value: string) => {
    let data = [...formData.data];
    const mappedData = data.map(e => {
      if (e.id === id) {
        if (isKey) e.key = value;
        else e.value = value;
      }
      return e;
    });
    setFormData((f: ExternalSecret) => ({ ...f, data: mappedData }));
  };

  return (
    <Flex gap="16" column wide>
      <Text size="medium" semiBold>
        <Switch
          color="primary"
          value={formData.includeAllProps}
          onChange={(evt, checked) =>
            setFormData((f: ExternalSecret) => ({
              ...f,
              includeAllProps: checked,
            }))
          }
        />
        Include all properties
      </Text>
      {!formData.includeAllProps && (
        <Flex column wide>
          {formData.data.map(obj => (
            <div key={obj.id} className="secret-data-list">
              <InputDebounced
                required
                name="dataSecretKey"
                label="PROPERTY"
                placeholder="Secret Property"
                value={obj.key}
                handleFormData={val => handleSecretChange(obj.id, true, val)}
                error={validateForm && !obj.key}
              />
              <InputDebounced
                name="dataSecretValue"
                label="Secret Key"
                placeholder="Secret key"
                value={obj.value}
                handleFormData={val => handleSecretChange(obj.id, false, val)}
                error={validateForm && !obj.value}
              />
              {formData.data.length > 1 && (
                <RemoveCircleOutlineIcon
                  className="remove-icon"
                  onClick={() =>
                    setFormData((f: ExternalSecret) => ({
                      ...f,
                      data: f.data.filter(e => e.id !== obj.id),
                    }))
                  }
                />
              )}
            </div>
          ))}
          <Button
            className="add-secret-data"
            startIcon={<Icon type={IconType.AddIcon} size="base" />}
            onClick={() =>
              setFormData((f: ExternalSecret) => ({
                ...f,
                data: [
                  ...f.data,
                  {
                    id:
                      f.data.length > 0 ? f.data[f.data.length - 1].id + 1 : 1,
                    key: '',
                    value: '',
                  },
                ],
              }))
            }
          >
            Add
          </Button>
        </Flex>
      )}
    </Flex>
  );
};
