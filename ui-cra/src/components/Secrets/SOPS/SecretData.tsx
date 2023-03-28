import RemoveCircleOutlineIcon from '@material-ui/icons/RemoveCircleOutline';
import { Button, Icon, IconType } from '@weaveworks/weave-gitops';
import { Dispatch } from 'react';
import InputDebounced from './InputDebounced';
import { SOPS } from './utils';

const data = ({
  formData,
  validateForm,
  setFormData,
}: {
  formData: SOPS;
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
    setFormData((f: SOPS) => ({ ...f, data: mappedData }));
  };

  return (
    <>
      {formData.data.map(obj => (
        <div key={obj.id} className="secret-data-list">
          <InputDebounced
            required
            name="dataSecretKey"
            label="KEY"
            placeholder="Secret key"
            value={obj.key}
            handleFormData={val => handleSecretChange(obj.id, true, val)}
            error={validateForm && !obj.key}
          />
          <InputDebounced
            required
            name="dataSecretValue"
            label="VALUE"
            placeholder="secret value"
            value={obj.value}
            handleFormData={val => handleSecretChange(obj.id, false, val)}
            error={validateForm && !obj.value}
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
