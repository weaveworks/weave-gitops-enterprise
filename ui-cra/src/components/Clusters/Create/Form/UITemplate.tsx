import { FieldTemplateProps, ObjectFieldTemplateProps } from '@rjsf/core';
import React from 'react';
import {
  DefaultObjectFieldTemplate,
  DefaultFieldTemplate,
} from './DefaultTemplates';
import styled from 'styled-components';

const FieldWrapper = styled.div`
  width: '100%';
`;

export function FieldTemplate(props: FieldTemplateProps) {
  const Template = props.uiSchema['ui:template'];
  if (
    Template &&
    props.schema.type !== 'object' &&
    props.schema.type !== 'array'
  ) {
    return <Template {...props} />;
  } else {
    return (
      <FieldWrapper>
        <DefaultFieldTemplate {...props} />
      </FieldWrapper>
    );
  }
}

export const ObjectFieldTemplate = defaultOrComponent(
  DefaultObjectFieldTemplate,
);

function defaultOrComponent(
  DefaultTemplate: React.FunctionComponent<ObjectFieldTemplateProps>,
) {
  return function (props: ObjectFieldTemplateProps) {
    const Template = props.uiSchema['ui:template'];
    if (Template) {
      return <Template {...props} />;
    } else {
      return <DefaultTemplate {...props} />;
    }
  };
}
