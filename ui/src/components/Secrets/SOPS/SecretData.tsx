import RemoveCircleOutlineIcon from '@material-ui/icons/RemoveCircleOutline';
import { Button, Icon, IconType } from '@weaveworks/weave-gitops';
import { Dispatch } from 'react';
import { InputDebounced } from '../../../utils/form';
import { ExternalSecret, SOPS } from '../Shared/utils';

const data = ({
  formData,
  setFormData,
  formError,
}: {
  formData: SOPS | ExternalSecret;
  setFormData: Dispatch<React.SetStateAction<any>>;
  formError: string;
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
            error={formError === 'data' && !obj.key}
          />
          <InputDebounced
            required
            name="dataSecretValue"
            label="VALUE"
            placeholder="secret value"
            value={obj.value}
            handleFormData={val => handleSecretChange(obj.id, false, val)}
            error={formError === 'data' && !obj.value}
          />
          {formData.data.length > 1 && (
            <RemoveCircleOutlineIcon
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
    </>
  );
};

export default data;
