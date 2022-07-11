import React, { FC, Dispatch } from 'react';
import { Button } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { Template } from '../../../../../cluster-services/cluster_services.pb';
import FormControl from '@material-ui/core/FormControl';
import Select from '@material-ui/core/Select';
import MenuItem from '@material-ui/core/MenuItem';
import Input from '@material-ui/core/Input';
import FormHelperText from '@material-ui/core/FormHelperText';

const FormWrapper = styled.form`
  .form-group {
    padding-top: ${({ theme }) => theme.spacing.base};
  }

  .preview-cta {
    display: flex;
    justify-content: flex-end;
    padding-top: ${({ theme }) => theme.spacing.small};
    padding-bottom: ${({ theme }) => theme.spacing.base};
    button {
      width: 200px;
    }
  }
`;

const TemplateFields: FC<{
  activeTemplate: Template | null;
  onPRPreview: () => void;
  onFormDataUpdate: Dispatch<React.SetStateAction<any>>;
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
}> = ({ activeTemplate, onPRPreview, formData, setFormData }) => {
  const handleFormData = (
    event:
      | React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>
      | React.ChangeEvent<{ name?: string; value: unknown }>,
  ) => {
    const { name, value } = event?.target;
    setFormData({ ...formData, [name as string]: value });
  };

  return (
    <FormWrapper>
      {activeTemplate?.parameters?.map(param => {
        const name = param.name || '';
        const options = param?.options || [];
        if (options.length > 0) {
          return (
            <FormControl style={{ width: '50%' }} id={name}>
              <span>{name}</span>
              <Select
                value={formData.name}
                onChange={handleFormData}
                autoWidth
                name={name}
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
                name={name}
                value={formData.name}
                onChange={handleFormData}
              />
              <FormHelperText>{param.description}</FormHelperText>
            </FormControl>
          );
      })}
      <div className="preview-cta">
        <Button onClick={onPRPreview}>PREVIEW PR</Button>
      </div>
    </FormWrapper>
  );
};

export default TemplateFields;
