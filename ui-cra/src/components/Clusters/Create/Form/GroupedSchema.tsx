// Adapted from : https://codesandbox.io/s/0y7787xp0l?file=/src/index.js:1507-1521

import { ObjectFieldTemplateProps } from '@rjsf/core';
import React, { ReactNode } from 'react';

export function ObjectFieldTemplate(props: ObjectFieldTemplateProps) {
  return (
    <>
      {doGrouping({
        formContext: props.formContext,
        properties: props.properties,
        groups: props.uiSchema['ui:groups'],
      })}
    </>
  );
}

const EXTRANEOUS = Symbol('EXTRANEOUS');
function doGrouping({
  properties,
  groups,
  formContext,
}: {
  properties: ObjectFieldTemplateProps['properties'];
  formContext: ObjectFieldTemplateProps['formContext'];
  groups: string | object;
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
        return el.content;
      }
      return EXTRANEOUS;
    } else if (typeof g === 'object') {
      const { templates, activeStep } = formContext;
      const GroupComponent = templates
        ? templates[g['ui:template']]
        : DefaultTemplate;
      const _properties = Object.keys(g).reduce(
        (
          acc: { name: string; active: boolean; children: ReactNode }[],
          key: string,
        ) => {
          const field = g[key];
          if (key.startsWith('ui:')) return acc;
          if (!Array.isArray(field)) return acc;
          return [
            ...acc,
            {
              name: key,
              active: key === activeStep,
              children: doGrouping({
                formContext,
                properties,
                groups: field,
              }),
            },
          ];
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
  return props?.properties?.map((p, index) => (
    <div key={index}>{p.content}</div>
  ));
}
