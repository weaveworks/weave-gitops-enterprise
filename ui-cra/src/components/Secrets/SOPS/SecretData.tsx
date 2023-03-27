import {
  FormControl,
  FormControlLabel,
  Radio,
  RadioGroup
} from '@material-ui/core';
import RemoveCircleOutlineIcon from '@material-ui/icons/RemoveCircleOutline';
import { Button, Icon, IconType } from '@weaveworks/weave-gitops';
import { Dispatch } from 'react';
import InputDebounced from './InputDebounced';
import { SecretDataType, SOPS } from './utils';

const data = ({
  formData,
  setFormData,
}: {
  formData: SOPS;
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
    setFormData((f: SOPS) => ({ ...f, data: mappedData }));
  };

  return (
    <>
      <div className="form-group">
        <FormControl>
          <RadioGroup
            row
            aria-labelledby="demo-controlled-radio-buttons-group"
            name="controlled-radio-buttons-group"
            value={formData.secretType}
            onChange={event =>
              setFormData((f: SOPS) => ({
                ...f,
                secretType: parseInt(event.target.value),
              }))
            }
          >
            <FormControlLabel
              value={SecretDataType.value}
              control={<Radio />}
              label="stringData"
            />
            <FormControlLabel
              value={SecretDataType.KeyValue}
              control={<Radio />}
              label="data"
            />
          </RadioGroup>
        </FormControl>
      </div>
      {formData.data.map((obj, index) => (
        <div key={obj.id} className="secret-data-list">
          <InputDebounced
            required
            name="dataSecretKey"
            label="KEY"
            placeholder="Secret key"
            value={obj.key}
            handleFormData={val => handleSecretChange(obj.id, true, val)}
          />
          <InputDebounced
            required
            name="dataSecretValue"
            label="VALUE"
            placeholder="secret value"
            value={obj.value}
            handleFormData={val => handleSecretChange(obj.id, false, val)}
          />
          <RemoveCircleOutlineIcon
            className="remove-icon"
            onClick={() =>
              setFormData((f: SOPS) => ({
                ...f,
                data: f.data.filter(e => e.id !== obj.id),
              }))
            }
          />
        </div>
      ))}
      <Button
        className="add-secret-data"
        startIcon={<Icon type={IconType.AddIcon} size="base" />}
        onClick={() =>
          setFormData((f: SOPS) => ({
            ...f,
            data: [
              ...f.data,
              {
                id: f.data.length > 0 ? f.data[f.data.length - 1].id + 1 : 1,
                key: '',
                value: '',
              },
            ],
          }))
        }
      >
        Add
      </Button>
    </>
  );
};

export default data;
