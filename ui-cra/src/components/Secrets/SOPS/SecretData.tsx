import {
  FormControl,
  FormControlLabel,
  Radio,
  RadioGroup,
} from '@material-ui/core';
import { Button, Icon, IconType } from '@weaveworks/weave-gitops';
import { useState } from 'react';
import { EncryptionType, SOPS } from '.';
import InputDebounced from './InputDebounced';
import RemoveCircleOutlineIcon from '@material-ui/icons/RemoveCircleOutline';

const SecretData = ({
  formData,
  handleFormData,
}: {
  formData: SOPS;
  handleFormData: (value: any, key: string) => void;
}) => {
  const [type, setType] = useState('stringData');
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setType((event.target as HTMLInputElement).value);
  };
  const handleSecretChange = (index: number, isKey: boolean, value: string) => {
    const mappedData = formData.secretData.map((e, i) => {
      if (i === index) {
        if (isKey) e.key = value;
        else e.value = value;
        return e;
      }
      return e;
    });
    return mappedData;
  };
  return (
    <div className="group-section">
      <h2>Secret Data</h2>
      <p className="secret-data-hint">
        Please note that we will encode the secret values to base64 before
        encryption
      </p>
      <div className="form-group">
        <FormControl>
          <RadioGroup
            row
            aria-labelledby="demo-controlled-radio-buttons-group"
            name="controlled-radio-buttons-group"
            value={formData.encryptionType}
            onChange={event =>
              handleFormData(event.target.value, 'encryptionType')
            }
          >
            <FormControlLabel
              value={EncryptionType.value}
              control={<Radio />}
              label="String Data"
            />
            <FormControlLabel
              value={EncryptionType.KeyValue}
              control={<Radio />}
              label="Data"
            />
          </RadioGroup>
        </FormControl>
      </div>
      {type === 'stringData' ? (
        <InputDebounced
          required
          name="secretValue"
          label="SECRET VALUE"
          value={formData.secretValue}
          handleFormData={val => handleFormData(val, 'secretValue')}
        />
      ) : (
        <>
          {formData.secretData.map((obj, index) => (
            <div key={index} className="secret-data-list">
              <InputDebounced
                required
                name="dataSecretKey"
                label="KEY"
                placeholder="Secret key"
                value={obj.key}
                handleFormData={val =>
                  handleFormData(
                    handleSecretChange(index, true, val),
                    'secretData',
                  )
                }
              />
              <InputDebounced
                required
                name="dataSecretValue"
                label="VALUE"
                placeholder="secret value"
                value={obj.value}
                handleFormData={val =>
                  handleFormData(
                    handleSecretChange(index, false, val),
                    'secretData',
                  )
                }
              />
              {formData.secretData.length > 1 && (
                <RemoveCircleOutlineIcon
                  className="remove-icon"
                  onClick={() => {
                    formData.secretData.splice(index, 1);
                    console.log(formData.secretData);
                    handleFormData([...formData.secretData], 'secretData');
                  }}
                />
              )}
            </div>
          ))}
          <Button
            className="add-secret-data"
            startIcon={<Icon type={IconType.AddIcon} size="base" />}
            onClick={() =>
              handleFormData(
                [...formData.secretData, { key: '', value: '' }],
                'secretData',
              )
            }
          >
            Add
          </Button>
        </>
      )}
    </div>
  );
};

export default SecretData;
