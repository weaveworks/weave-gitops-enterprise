import React, { Dispatch, FC } from 'react';
import { useRouteMatch } from 'react-router-dom';
import styled from 'styled-components';
import { TemplateEnriched } from '../../../../types/custom';
import { Input, Select } from '../../../../utils/form';
import { Routes } from '../../../../utils/nav';

const TemplateFieldsWrapper = styled.div`
  div[class*='MuiFormControl-root'] {
    width: calc(100% / 3);
  }
`;

const TemplateFields: FC<{
  template: TemplateEnriched;
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
  formError: string;
}> = ({ template, formData, setFormData, formError }) => {
  const UNEDITABLE_FIELDS = template.parameters
    ?.filter(param => Boolean(param.editable))
    .map(param => param.name);
  const { isExact: isEditing } = useRouteMatch(Routes.EditResource) || {};
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
    <TemplateFieldsWrapper>
      {template.parameters?.map((param, index) => {
        const name = param.name || '';
        const options = param?.options || [];
        const required = Boolean(!param.default && param.required);
        if (options.length > 0) {
          return (
            <Select
              key={index}
              name={name}
              required={required}
              label={name}
              value={parameterValues[name] || param.default}
              onChange={event => handleFormData(event, name)}
              items={options}
              description={param.description}
              error={formError === name && !parameterValues[name]}
            />
          );
        } else {
          return (
            <Input
              key={index}
              required={required}
              name={name}
              label={name}
              value={parameterValues[name]}
              placeholder={param.default}
              onChange={handleFormData}
              description={param.description}
              disabled={isEditing && UNEDITABLE_FIELDS?.includes(name)}
              error={formError === name && !parameterValues[name]}
            />
          );
        }
      })}
    </TemplateFieldsWrapper>
  );
};

export default TemplateFields;
