import React, {
  FC,
  useMemo,
  useState,
  Dispatch,
  useEffect,
  useCallback,
  ReactNode,
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

const EXTRANEOUS = Symbol('EXTRANEOUS');

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
  onProfilesUpdate: Dispatch<React.SetStateAction<UpdatedProfile[]>>;
  onFormDataUpdate: Dispatch<React.SetStateAction<any>>;
  formData: any;
}> = ({
  activeTemplate,
  onPRPreview,
  activeStep,
  setActiveStep,
  clickedStep,
  onProfilesUpdate,
  formData,
  onFormDataUpdate,
}) => {
  const { updatedProfiles } = useProfiles();
  const [selectedProfiles, setSelectedProfiles] = useState<UpdatedProfile[]>(
    [],
  );
  const [userSelectedFields, setUserSelectedFields] = useState<string[]>([]);
  const [formContextId, setFormContextId] = useState<number>(0);

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

  const [uiSchema, setuiSchema] = useState<any>({
    'ui:groups': sections,
    'ui:template': (props: ObjectFieldTemplateProps) => (
      <ObjectFieldTemplate {...props} />
    ),
  });

  const handleSelectProfiles = useCallback(
    (profiles: UpdatedProfile[]) => {
      setSelectedProfiles(profiles);
      onProfilesUpdate(profiles);
    },
    [onProfilesUpdate],
  );

  const DefaultTemplate = useCallback(
    (props: { properties: ObjectFieldTemplateProps['properties'] }) => {
      return props?.properties?.map((p, index) => (
        <div key={index}>{p.content}</div>
      ));
    },
    [],
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

  const doGrouping = useCallback(
    ({
      properties,
      groups,
      formContext,
      previouslyVisibleFields,
      userSelectedFields,
    }: {
      properties: ObjectFieldTemplateProps['properties'];
      formContext: ObjectFieldTemplateProps['formContext'];
      groups: string | object;
      previouslyVisibleFields: string[];
      userSelectedFields: string[];
    }) => {
      if (!Array.isArray(groups)) {
        return properties?.map((property, index) => {
          return <div key={index}>{property.content}</div>;
        });
      }
      const mapped = groups.map((g, index) => {
        if (typeof g === 'string') {
          const found = properties?.filter(property => property.name === g);
          if (found?.length === 1) {
            const el = found[0];

            const firstOfAKind = previouslyVisibleFields.includes(
              el.content.props.name,
            )
              ? false
              : true;
            let visible =
              firstOfAKind ||
              userSelectedFields.includes(el.content.props.name);

            return React.cloneElement(el.content, { visible, firstOfAKind });
          }
          return EXTRANEOUS;
        } else if (typeof g === 'object') {
          const { templates, activeStep, setActiveStep, clickedStep } =
            formContext;
          const GroupComponent = templates
            ? templates[g['ui:template']]
            : DefaultTemplate;

          let previouslyVisibleFields: string[] = [];

          const _properties = Object.keys(g).reduce(
            (
              acc: {
                name: string;
                active: boolean;
                clicked: boolean;
                setActiveStep: Dispatch<
                  React.SetStateAction<string | undefined>
                >;
                children: ReactNode;
              }[],
              key: string,
            ) => {
              const field = g[key];

              if (key.startsWith('ui:')) return acc;
              if (!Array.isArray(field)) return acc;

              const section = [
                ...acc,
                {
                  name: key,
                  active: key === activeStep,
                  clicked: key === clickedStep,
                  setActiveStep,
                  addUserSelectedFields,
                  children: doGrouping({
                    formContext,
                    properties,
                    groups: field,
                    previouslyVisibleFields,
                    userSelectedFields,
                  }),
                },
              ];

              previouslyVisibleFields = Array.from(
                new Set([...previouslyVisibleFields, ...field]),
              );

              return section;
            },
            [],
          );

          return <GroupComponent key={index} properties={_properties} />;
        }

        throw new Error('Invalid object type: ' + typeof g + ' ' + g);
      });

      return mapped;
    },
    [DefaultTemplate, addUserSelectedFields],
  );

  const ObjectFieldTemplate = useCallback(
    (props: ObjectFieldTemplateProps) => {
      return (
        <>
          {doGrouping({
            formContext: props.formContext,
            properties: props.properties,
            groups: props.uiSchema['ui:groups'],
            previouslyVisibleFields: [],
            userSelectedFields,
          })}
        </>
      );
    },
    [doGrouping, userSelectedFields],
  );

  useEffect(() => {
    setFormContextId((prevState: any) => prevState + 1);

    setuiSchema({
      'ui:groups': sections,
      'ui:template': (props: ObjectFieldTemplateProps) => (
        <ObjectFieldTemplate {...props} />
      ),
    });

    const requiredProfiles = updatedProfiles.filter(
      profile => profile.required === true,
    );
    handleSelectProfiles(requiredProfiles);
  }, [
    updatedProfiles,
    onProfilesUpdate,
    handleSelectProfiles,
    sections,
    ObjectFieldTemplate,
  ]);

  return (
    <FormWrapper
      schema={schema as JSONSchema7}
      onChange={({ formData }) => onFormDataUpdate(formData)}
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
