// Adapted from : https://codesandbox.io/s/0y7787xp0l?file=/src/index.js:1507-1521

import { ObjectFieldTemplateProps } from '@rjsf/core';
import React, { Dispatch, ReactNode } from 'react';

export function ObjectFieldTemplate(props: ObjectFieldTemplateProps) {
  return (
    <>
      {doGrouping({
        formContext: props.formContext,
        properties: props.properties,
        groups: props.uiSchema['ui:groups'],
        previouslyVisibleFields: [],
      })}
    </>
  );
}

const EXTRANEOUS = Symbol('EXTRANEOUS');

function doGrouping({
  properties,
  groups,
  formContext,
  previouslyVisibleFields,
}: {
  properties: ObjectFieldTemplateProps['properties'];
  formContext: ObjectFieldTemplateProps['formContext'];
  groups: string | object;
  previouslyVisibleFields: string[];
}) {
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

        const visible = previouslyVisibleFields.includes(el.content.props.name)
          ? false
          : true;

        return React.cloneElement(el.content, { visible });
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
              children: doGrouping({
                formContext,
                properties,
                groups: field,
                previouslyVisibleFields,
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
}

function DefaultTemplate(props: {
  properties: ObjectFieldTemplateProps['properties'];
}) {
  return props?.properties?.map((p, index) => {
    return <div key={index}>{p.content}</div>;
  });
}
