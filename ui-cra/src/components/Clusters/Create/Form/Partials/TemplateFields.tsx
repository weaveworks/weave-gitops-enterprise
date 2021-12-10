import React, {
  FC,
  useMemo,
  useState,
  Dispatch,
  useEffect,
  useCallback,
} from 'react';
import weaveTheme from 'weaveworks-ui-components/lib/theme';
import { Button } from 'weaveworks-ui-components';
import {
  Template,
  TemplateObject,
  UpdatedProfile,
} from '../../../../../types/custom';
import { ObjectFieldTemplateProps } from '@rjsf/core';
import { JSONSchema7 } from 'json-schema';
import Form from '@rjsf/material-ui';
import * as UiTemplate from '../UITemplate';
import FormSteps from '../Steps';
import MultiSelectDropdown from '../../../../MultiSelectDropdown';
import ProfilesList from './ProfilesList';
import useProfiles from '../../../../../contexts/Profiles';
import { FormStep } from '../Step';
import styled from 'styled-components';
import ObjectFieldTemplate from '../GroupedSchema';

const base = weaveTheme.spacing.base;
const small = weaveTheme.spacing.small;

const FormWrapper = styled(Form)`
  .form-group {
    padding-top: ${base};
  }
  .profiles {
    .profiles-select {
      display: flex;
      align-items: center;
    }
    .previewCTA {
      display: flex;
      justify-content: flex-end;
      padding-top: ${small};
      padding-bottom: ${base};
    }
  }
`;

const TemplateFields: FC<{
  activeTemplate: Template | null;
  onPRPreview: () => void;
  activeStep: string | undefined;
  setActiveStep: Dispatch<React.SetStateAction<string | undefined>>;
  clickedStep: string;
  onFormDataUpdate: Dispatch<React.SetStateAction<any>>;
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
}> = ({
  activeTemplate,
  onPRPreview,
  activeStep,
  setActiveStep,
  clickedStep,
  // onProfilesUpdate,
  formData,
  setFormData,
}) => {
  const { updatedProfiles } = useProfiles();
  const [selectedProfiles, setSelectedProfiles] = useState<UpdatedProfile[]>(
    [],
  );
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

  const handleSelectProfiles = useCallback(
    (profiles: UpdatedProfile[]) => {
      setSelectedProfiles(profiles);
      setFormData((prevState: any) => ({ ...prevState, profiles }));
    },
    [setFormData],
  );

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

  console.log('formData', formData);

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

    const requiredProfiles = updatedProfiles.filter(
      profile => profile.required === true,
    );

    console.log(requiredProfiles, formData.profiles);

    if (formData?.profiles) {
      handleSelectProfiles(formData?.profiles);
    } else {
      handleSelectProfiles(requiredProfiles);
    }
  }, [
    updatedProfiles,
    handleSelectProfiles,
    sections,
    addUserSelectedFields,
    userSelectedFields,
    formData.profiles,
  ]);

  // how is the form updated from the form data that comes from callbackState

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
      <FormStep
        className="profiles"
        title="Profiles"
        active={activeStep === 'Profiles'}
        clicked={clickedStep === 'Profiles'}
        setActiveStep={setActiveStep}
      >
        <div className="profiles-select">
          <span>Select profiles:&nbsp;</span>
          <MultiSelectDropdown
            items={updatedProfiles}
            onSelectItems={handleSelectProfiles}
          />
        </div>
        <ProfilesList
          selectedProfiles={selectedProfiles}
          onProfilesUpdate={handleSelectProfiles}
        />
        <div className="previewCTA">
          <Button>Preview PR</Button>
        </div>
      </FormStep>
    </FormWrapper>
  );
};

export default TemplateFields;
