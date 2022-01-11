import React, {
  FC,
  useMemo,
  useState,
  Dispatch,
  useEffect,
  useCallback,
} from 'react';
import weaveTheme from 'weaveworks-ui-components/lib/theme';
import { Template, TemplateObject } from '../../../../../types/custom';
import { ObjectFieldTemplateProps } from '@rjsf/core';
import { JSONSchema7 } from 'json-schema';
import Form from '@rjsf/material-ui';
import * as UiTemplate from '../UITemplate';
import FormSteps from '../Steps';
import styled from 'styled-components';
import ObjectFieldTemplate from '../GroupedSchema';
import { Button } from 'weaveworks-ui-components';

const base = weaveTheme.spacing.base;
const small = weaveTheme.spacing.small;

const FormWrapper = styled(Form)`
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
  setActiveStep: Dispatch<React.SetStateAction<string | undefined>>;
  clickedStep: string;
}> = ({
  activeTemplate,
  onPRPreview,
  formData,
  setFormData,
  setActiveStep,
  clickedStep,
}) => {
  const [userSelectedFields, setUserSelectedFields] = useState<string[]>([]);
  const [formContextId, setFormContextId] = useState<number>(0);
  const [uiSchema, setuiSchema] = useState<any>({});

  const objectTitle = (object: TemplateObject, index: number) => {
    if (object.displayName && object.displayName !== '') {
      return `${index + 1}.${object.kind} (${object.displayName})`;
    }
    return `${index + 1}.${object.kind}`;
  };

  const required = useMemo(() => {
    return activeTemplate?.parameters?.map(param => param.name);
  }, [activeTemplate]);

  const parameters = useMemo(() => {
    return (
      activeTemplate?.parameters?.map(param => {
        const { name, options } = param;
        if (options?.length !== 0) {
          return {
            [name]: {
              type: 'string',
              title: `${name}`,
              enum: options,
            },
          };
        } else {
          return {
            [name]: {
              type: 'string',
              title: `${name}`,
            },
          };
        }
      }) || []
    );
  }, [activeTemplate]);

  const properties = useMemo(() => {
    return Object.assign({}, ...parameters);
  }, [parameters]);

  const schema: JSONSchema7 = useMemo(() => {
    return {
      type: 'object',
      properties,
      required,
    };
  }, [properties, required]);

  // Adapted from : https://codesandbox.io/s/0y7787xp0l?file=/src/index.js:1507-1521
  const sections = useMemo(() => {
    const excludeObjectsWithoutParams = activeTemplate?.objects?.filter(
      object => object.parameters?.length !== 0,
    );
    const groups =
      excludeObjectsWithoutParams?.reduce(
        (accumulator, item, index) =>
          Object.assign(accumulator, {
            [objectTitle(item, index)]: item.parameters,
          }),
        {},
      ) || {};
    Object.assign(groups, { 'ui:template': 'Box' });
    return [groups];
  }, [activeTemplate]);

  const addUserSelectedFields = useCallback(
    (name: string) => {
      if (userSelectedFields.includes(name)) {
        setUserSelectedFields(userSelectedFields.filter(el => el !== name));
      } else {
        setUserSelectedFields([...userSelectedFields, name]);
      }
    },
    [userSelectedFields],
  );

  useEffect(() => {
    setFormContextId((prevState: number) => prevState + 1);

    setuiSchema({
      'ui:groups': sections,
      'ui:template': (props: ObjectFieldTemplateProps) => (
        <ObjectFieldTemplate
          {...props}
          userSelectedFields={userSelectedFields}
          addUserSelectedFields={addUserSelectedFields}
        />
      ),
    });
  }, [sections, addUserSelectedFields, userSelectedFields]);

  return (
    <FormWrapper
      schema={schema as JSONSchema7}
      onChange={({ formData }) => setFormData(formData)}
      formData={formData}
      onSubmit={onPRPreview}
      onError={() => console.log('errors')}
      uiSchema={uiSchema}
      formContext={{
        formContextId,
        templates: FormSteps,
        clickedStep,
        setActiveStep,
      }}
      {...UiTemplate}
    >
      <div className="previewCTA">
        <Button>Preview PR</Button>
      </div>
    </FormWrapper>
  );
};

export default TemplateFields;
