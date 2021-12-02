// Adapted from : https://codesandbox.io/s/0y7787xp0l?file=/src/index.js:1507-1521

import { ObjectFieldTemplateProps } from '@rjsf/core';
import React, { Dispatch, ReactNode } from 'react';

const EXTRANEOUS = Symbol('EXTRANEOUS');

const DefaultTemplate = (props: {
  properties: ObjectFieldTemplateProps['properties'];
}) => {
  return props?.properties?.map((p, index) => (
    <div key={index}>{p.content}</div>
  ));
};

const doGrouping = ({
  properties,
  groups,
  formContext,
  previouslyVisibleFields,
  userSelectedFields,
  addUserSelectedFields,
}: {
  properties: ObjectFieldTemplateProps['properties'];
  formContext: ObjectFieldTemplateProps['formContext'];
  groups: string | object;
  previouslyVisibleFields: string[];
  userSelectedFields: string[];
  addUserSelectedFields: (name: string) => void;
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
          firstOfAKind || userSelectedFields.includes(el.content.props.name);
        let disabled = !firstOfAKind;

        return React.cloneElement(el.content, {
          visible,
          firstOfAKind,
          disabled,
        });
      }
      return EXTRANEOUS;
    } else if (typeof g === 'object') {
      const { templates, activeStep, setActiveStep, clickedStep } = formContext;
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
            setActiveStep: Dispatch<React.SetStateAction<string | undefined>>;
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
                addUserSelectedFields,
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
};

const ObjectFieldTemplate = (props: any) => {
  return (
    <>
      {doGrouping({
        formContext: props.formContext,
        properties: props.properties,
        groups: props.uiSchema['ui:groups'],
        previouslyVisibleFields: [],
        userSelectedFields: props.userSelectedFields,
        addUserSelectedFields: props.addUserSelectedFields,
      })}
    </>
  );
};

export default ObjectFieldTemplate;
