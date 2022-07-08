import React, { FC, Dispatch } from 'react';
import { Button } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { Template } from '../../../../../cluster-services/cluster_services.pb';
import FormControl from '@material-ui/core/FormControl';
import Select from '@material-ui/core/Select';
import MenuItem from '@material-ui/core/MenuItem';
import Input from '@material-ui/core/Input';
import FormHelperText from '@material-ui/core/FormHelperText';

const FormWrapper = styled.div`
  .form-group {
    padding-top: ${({ theme }) => theme.spacing.base};
  }

  .previewCTA {
    display: flex;
    justify-content: flex-end;
    padding-top: ${({ theme }) => theme.spacing.small};
    padding-bottom: ${({ theme }) => theme.spacing.base};
  }
`;

const TemplateFields: FC<{
  activeTemplate: Template | null;
  onPRPreview: () => void;
  onFormDataUpdate: Dispatch<React.SetStateAction<any>>;
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
}> = ({ activeTemplate, onPRPreview, formData, setFormData }) => {
  console.log(formData);
  const handleFormData = (event: any) => {
    const id = event?.target?.id;
    const value = event?.target?.value;
    setFormData({ ...formData, [id]: value });
  };

  return (
    <FormWrapper>
      {activeTemplate?.parameters?.map(param => {
        const name = param.name || '';
        const options = param?.options || [];
        if (options.length > 0) {
          return (
            <FormControl style={{ width: '50%' }}>
              <span>{name}</span>
              <Select
                id={name}
                value={formData.name}
                onChange={handleFormData}
                autoWidth
                label={name}
              >
                {options?.map(option => (
                  <MenuItem key={option} value={option}>
                    {option}
                  </MenuItem>
                ))}
              </Select>
              <FormHelperText>{param.description}</FormHelperText>
            </FormControl>
          );
        } else
          return (
            <FormControl style={{ width: '50%' }}>
              <span>{name}</span>
              <Input
                id={name}
                value={formData.name}
                onChange={handleFormData}
              />
              <FormHelperText>{param.description}</FormHelperText>
            </FormControl>
          );
      })}
      <div className="previewCTA">
        {/* add handler for pr preview */}
        <Button type="submit">PREVIEW PR</Button>
      </div>
    </FormWrapper>
  );
};

export default TemplateFields;
