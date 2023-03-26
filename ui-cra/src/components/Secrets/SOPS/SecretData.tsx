import {
  FormControl,
  FormControlLabel,
  Radio,
  RadioGroup,
} from '@material-ui/core';
import { Button, Icon, IconType } from '@weaveworks/weave-gitops';
import { SecretDataType, SOPS } from './utils';
import InputDebounced from './InputDebounced';
import RemoveCircleOutlineIcon from '@material-ui/icons/RemoveCircleOutline';

const data = ({
  formData,
  handleFormData,
}: {
  formData: SOPS;
  handleFormData: (value: any, key: string) => void;
}) => {
  const handleSecretChange = (index: number, isKey: boolean, value: string) => {
    const mappedData = formData.data.map((e, i) => {
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
    <>
      <div className="form-group">
        <FormControl>
          <RadioGroup
            row
            aria-labelledby="demo-controlled-radio-buttons-group"
            name="controlled-radio-buttons-group"
            value={formData.secretType}
            onChange={event =>
              handleFormData(parseInt(event.target.value), 'secretType')
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
        <div key={`${obj.key}-${index}`} className="secret-data-list">
          <InputDebounced
            required
            name="dataSecretKey"
            label="KEY"
            placeholder="Secret key"
            value={obj.key}
            handleFormData={val =>
              handleFormData(handleSecretChange(index, true, val), 'data')
            }
          />
          <InputDebounced
            required
            name="dataSecretValue"
            label="VALUE"
            placeholder="secret value"
            value={obj.value}
            handleFormData={val =>
              handleFormData(handleSecretChange(index, false, val), 'data')
            }
          />
          <RemoveCircleOutlineIcon
            className="remove-icon"
            onClick={() => {
              formData.data.splice(index, 1);
              handleFormData([...formData.data], 'data');
            }}
          />
        </div>
      ))}

      <Button
        className="add-secret-data"
        startIcon={<Icon type={IconType.AddIcon} size="base" />}
        onClick={() =>
          handleFormData([...formData.data, { key: '', value: '' }], 'data')
        }
      >
        Add
      </Button>
    </>
  );
};

export default data;
