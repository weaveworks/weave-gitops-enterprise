import { Switch } from '@material-ui/core';
import RemoveCircleOutlineIcon from '@material-ui/icons/RemoveCircleOutline';
import { Button, Flex, Icon, IconType, Text } from '@weaveworks/weave-gitops';
import { Dispatch } from 'react';
import { InputDebounced } from '../../../utils/form';
import { ExternalSecret } from '../Shared/utils';
import styled from 'styled-components';

const PropertiesSwitch = styled(Switch)`
  .MuiSwitch-track {
    background-color: ${({ theme }) => theme.colors.primary30};
  }
`;

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
    setFormData((f: ExternalSecret) => ({
      ...f,
      data: f.data.map(p => {
        if (p.id !== id) return p;

        if (isKey) p.key = value;
        else p.value = value;

        return p;
      }),
    }));
  };

  const handleRemoveProp = (id: number) => {
    setFormData((f: ExternalSecret) => ({
      ...f,
      data: f.data.filter(e => e.id !== id),
    }));
  };
  const handleNewProp = () => {
    setFormData((f: ExternalSecret) => ({
      ...f,
      data: [
        ...f.data,
        {
          id: formData.data[formData.data.length - 1].id + 1,
          key: '',
          value: '',
        },
      ],
    }));
  };
  return (
    <Flex gap="16" column wide>
      <Text size="medium" semiBold>
        <PropertiesSwitch
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
              />
              {formData.data.length > 1 && (
                <RemoveCircleOutlineIcon
                  style={{ marginRight: '-30px' }}
                  className="remove-icon"
                  onClick={() => handleRemoveProp(obj.id)}
                />
              )}
            </div>
          ))}
          <Button
            className="add-secret-data"
            startIcon={<Icon type={IconType.AddIcon} size="base" />}
            onClick={() => handleNewProp()}
          >
            Add
          </Button>
        </Flex>
      )}
    </Flex>
  );
};
