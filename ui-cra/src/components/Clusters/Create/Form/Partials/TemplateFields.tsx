import React, { FC, Dispatch } from 'react';
import { Button, LoadingPage } from '@weaveworks/weave-gitops';
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
  .preview-loading {
    padding: ${({ theme }) => theme.spacing.base};
  }
`;

const TemplateFields: FC<{
  activeTemplate: Template | null;
  onPRPreview: () => void;
  onFormDataUpdate: Dispatch<React.SetStateAction<any>>;
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
  previewLoading: boolean;
}> = ({
  activeTemplate,
  onPRPreview,
  formData,
  setFormData,
  previewLoading,
}) => {
  const handleFormData = (
    event:
      | React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>
      | React.ChangeEvent<{ name?: string; value: unknown }>,
    fieldName?: string,
  ) => {
    const { name, value } = event?.target;
    setFormData({ ...formData, [(name || fieldName) as string]: value });
  };

  return (
    <FormWrapper>
      {activeTemplate?.parameters?.map((param, index) => {
        const name = param.name || '';
        const options = param?.options || [];
        if (options.length > 0) {
          return (
            <Select
              key={index}
              className="form-section"
              name={name}
              required={param.required}
              label={name}
              value={formData[name]}
              onChange={event => handleFormData(event, name)}
              items={options}
              description={param.description}
            />
          );
        } else
          return (
            <Input
              key={index}
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
      {previewLoading ? (
        <LoadingPage className="preview-loading" />
      ) : (
        <div className="preview-cta">
          <Button onClick={event => validateFormData(event, onPRPreview)}>
            PREVIEW PR
          </Button>
        </div>
      )}
    </FormWrapper>
  );
};

export default TemplateFields;
