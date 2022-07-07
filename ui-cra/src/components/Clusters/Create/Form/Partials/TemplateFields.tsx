import React, { FC, Dispatch } from 'react';
import { theme as weaveTheme, Button } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import {
  Template,
  TemplateObject,
} from '../../../../../cluster-services/cluster_services.pb';
import FormControl from '@material-ui/core/FormControl';
import Select from '@material-ui/core/Select';
import MenuItem from '@material-ui/core/MenuItem';
import Input from '@material-ui/core/Input';
import FormHelperText from '@material-ui/core/FormHelperText';

const base = weaveTheme.spacing.base;
const small = weaveTheme.spacing.small;

const FormWrapper = styled.div`
  .form-group {
    padding-top: ${base};
  }

  .previewCTA {
    display: flex;
    justify-content: flex-end;
    padding-top: ${small};
    padding-bottom: ${base};
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
  const handleFormData = () => {
    // update formdata
  };

  return (
    <FormWrapper>
      {activeTemplate?.parameters?.map(param => {
        const name = param.name || '';
        const options = param?.options || [];
        if (options.length > 0) {
          return (
            <div>
              <FormControl>
                <Select
                  id={param.name}
                  value={formData.name}
                  onChange={handleFormData}
                  autoWidth
                  label={param.name}
                >
                  {options?.map(option => (
                    <MenuItem key={option} value={option}>
                      {option}
                    </MenuItem>
                  ))}
                </Select>
                <FormHelperText>{param.description}</FormHelperText>
              </FormControl>
            </div>
          );
        } else
          return (
            <div>
              <FormControl>
                <Input
                  id={param.name}
                  value={formData.name}
                  onChange={handleFormData}
                />
                <FormHelperText>{param.description}</FormHelperText>
              </FormControl>
            </div>
          );
      })}
      <div className="previewCTA">
        <Button type="submit">PREVIEW PR</Button>
      </div>
    </FormWrapper>
  );
};

export default TemplateFields;
