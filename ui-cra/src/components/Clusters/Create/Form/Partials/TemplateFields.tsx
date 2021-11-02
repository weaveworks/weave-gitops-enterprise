import React, { FC, useCallback, useMemo, useState, Dispatch } from 'react';
import weaveTheme from 'weaveworks-ui-components/lib/theme';
import { Button } from 'weaveworks-ui-components';
import { makeStyles, createStyles } from '@material-ui/core/styles';
import {
  Profile,
  Template,
  TemplateObject,
  UpdatedProfile,
} from '../../../../../types/custom';
import { ISubmitEvent, ObjectFieldTemplateProps } from '@rjsf/core';
import { JSONSchema7 } from 'json-schema';
import Form from '@rjsf/material-ui';
import * as Grouped from '../GroupedSchema';
import * as UiTemplate from '../UITemplate';
import FormSteps, { FormStep } from '../Steps';
import MultiSelectDropdown from '../../../../MultiSelectDropdown';
import ProfilesList from './ProfilesList';
import useProfiles from '../../../../../contexts/Profiles';
import useTemplates from '../../../../../contexts/Templates';

const base = weaveTheme.spacing.base;
const small = weaveTheme.spacing.small;

const useStyles = makeStyles(() =>
  createStyles({
    form: {
      paddingTop: base,
    },
    create: {
      paddingTop: small,
    },
    previewCTA: {
      display: 'flex',
      justifyContent: 'flex-end',
      paddingTop: small,
      paddingBottom: base,
    },
  }),
);

const TemplateFields: FC<{
  activeTemplate: Template | null;
  onSubmit: (formData: any, encodedProfiles: UpdatedProfile[]) => void;
  activeStep: string | undefined;
  setActiveStep: Dispatch<React.SetStateAction<string | undefined>>;
  clickedStep: string;
}> = ({ activeTemplate, onSubmit, activeStep, setActiveStep, clickedStep }) => {
  const classes = useStyles();
  const { profiles } = useProfiles();
  const { PRPreview } = useTemplates();
  const [formData, setFormData] = useState({});
  const [selectedProfiles, setSelectedProfiles] = useState<Profile[]>([]);
  const [updatedProfiles, setUpdatedProfiles] = useState<UpdatedProfile[]>([]);

  const objectTitle = (object: TemplateObject, index: number) => {
    if (object.displayName && object.displayName !== '') {
      return `${index + 1}.${object.kind} (${object.displayName})`;
    }
    return `${index + 1}.${object.kind}`;
  };

  const encodedProfiles = useCallback(
    (profiles: UpdatedProfile[]) =>
      profiles?.map(profile => {
        return {
          name: profile.name,
          version: profile.version,
          values: btoa(profile.values),
        };
      }),
    [],
  );

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
    const groups =
      activeTemplate?.objects?.reduce(
        (accumulator, item, index) =>
          Object.assign(accumulator, {
            [objectTitle(item, index)]: item.parameters,
          }),
        {},
      ) || {};
    Object.assign(groups, { 'ui:template': 'box' });
    return [groups];
  }, [activeTemplate]);

  const uiSchema = useMemo(() => {
    return {
      'ui:groups': sections,
      'ui:template': (props: ObjectFieldTemplateProps) => (
        <Grouped.ObjectFieldTemplate {...props} />
      ),
    };
  }, [sections]);

  const handleSubmit = useCallback(
    (event: ISubmitEvent<any>) => {
      setFormData(event.formData);
      onSubmit(event.formData, encodedProfiles(updatedProfiles));
    },
    [encodedProfiles, onSubmit, updatedProfiles],
  );

  const handleProfilesUpdate = useCallback(
    (updatedProfiles: UpdatedProfile[]) => {
      setUpdatedProfiles(updatedProfiles);
      PRPreview && onSubmit(formData, encodedProfiles(updatedProfiles));
    },
    [encodedProfiles, onSubmit, PRPreview, formData],
  );

  console.log(updatedProfiles);

  return useMemo(() => {
    return (
      <Form
        className={classes.form}
        schema={schema as JSONSchema7}
        onChange={({ formData }) => {
          console.log(formData);
          setFormData(formData);
          // only resubmit data if changes are made after the initial submit / the PR Preview is first generated
          PRPreview && onSubmit(formData, encodedProfiles(updatedProfiles));
        }}
        formData={formData}
        onSubmit={handleSubmit}
        onError={() => console.log('errors')}
        uiSchema={uiSchema}
        formContext={{
          templates: FormSteps,
          clickedStep,
          setActiveStep,
        }}
        {...UiTemplate}
      >
        <FormStep
          title="Profiles"
          active={activeStep === 'Profiles'}
          clicked={clickedStep === 'Profiles'}
          setActiveStep={setActiveStep}
        >
          <div style={{ display: 'flex', alignItems: 'center' }}>
            <span>Select profiles:&nbsp;</span>
            <MultiSelectDropdown
              items={profiles}
              onSelectItems={setSelectedProfiles}
            />
          </div>
          <ProfilesList
            selectedProfiles={selectedProfiles}
            onProfilesUpdate={handleProfilesUpdate}
          />
          <div className={classes.previewCTA}>
            <Button>Preview PR</Button>
          </div>
        </FormStep>
      </Form>
    );
  }, [
    classes,
    formData,
    schema,
    uiSchema,
    activeStep,
    setActiveStep,
    clickedStep,
    profiles,
    selectedProfiles,
    handleSubmit,
    encodedProfiles,
    onSubmit,
    updatedProfiles,
    PRPreview,
    handleProfilesUpdate,
  ]);
};

export default TemplateFields;
