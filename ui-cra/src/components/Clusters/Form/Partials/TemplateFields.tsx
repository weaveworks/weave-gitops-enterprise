import { Button, LoadingPage } from '@weaveworks/weave-gitops';
import React, { Dispatch, FC } from 'react';
import styled from 'styled-components';
import { TemplateEnriched } from '../../../../types/custom';
import { Input, Select, validateFormData } from '../../../../utils/form';
import { useLocation } from 'react-router-dom';

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
  template: TemplateEnriched;
  onPRPreview: () => void;
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
  previewLoading: boolean;
}> = ({ template, onPRPreview, formData, setFormData, previewLoading }) => {
  const UNEDITABLE_FIELDS = ['CLUSTER_NAME', 'NAMESPACE'];
  const { isExact: isEditing } =
    useRouteMatch(EDIT_CLUSTER_ROUTE) || {};
  const parameterValues = formData.parameterValues || {};
  const handleFormData = (
    event:
      | React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>
      | React.ChangeEvent<{ name?: string; value: unknown }>,
    fieldName?: string,
  ) => {
    const { name, value } = event?.target;
    setFormData({
      ...formData,
      parameterValues: {
        ...formData.parameterValues,
        [(name || fieldName) as string]: value,
      },
    });
  };

  return (
    <FormWrapper>
      {template.parameters?.map((param, index) => {
        const name = param.name || '';
        const options = param?.options || [];
        const required = Boolean(!param.default && param.required);
        const isEditing = location.pathname.includes('edit');
        if (options.length > 0) {
          return (
            <Select
              key={index}
              className="form-section"
              name={name}
              required={required}
              label={name}
              value={parameterValues[name] || param.default}
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
              required={required}
              name={name}
              label={name}
              value={parameterValues[name]}
              placeholder={param.default}
              onChange={handleFormData}
              description={param.description}
              disabled={isEditing && UNEDITABLE_FIELDS.includes(name)}
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
