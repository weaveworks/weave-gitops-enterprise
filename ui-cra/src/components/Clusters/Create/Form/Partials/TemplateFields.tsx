import React, { FC, Dispatch } from 'react';
import { Button } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { Template } from '../../../../../cluster-services/cluster_services.pb';
import { Input, Select, validateFormData } from '../../../../../utils/form';

const FormWrapper = styled.form`
  .form-section {
    width: 50%;
  }
  .preview-cta {
    display: flex;
    justify-content: flex-end;
    padding: ${({ theme }) => theme.spacing.small}
      ${({ theme }) => theme.spacing.base};
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
            <Select
              className="form-section"
              name={name}
              required={param.required}
              label={name}
              value={formData[name]}
              onChange={handleFormData}
              items={options}
              description={param.description}
            />
          );
        } else
          return (
            <Input
              className="form-section"
              required={param.required}
              name={name}
              label={name}
              value={formData[name]}
              onChange={handleFormData}
              description={param.description}
            />
          );
      })}
      <div className="preview-cta">
        <Button onClick={event => validateFormData(event, onPRPreview)}>
          PREVIEW PR
        </Button>
      </div>
    </FormWrapper>
  );
};

export default TemplateFields;
