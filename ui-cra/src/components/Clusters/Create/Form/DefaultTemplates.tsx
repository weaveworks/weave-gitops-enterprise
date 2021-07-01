import { FieldTemplateProps, ObjectFieldTemplateProps } from '@rjsf/core';
import React from 'react';

export function DefaultFieldTemplate(props: FieldTemplateProps) {
  const { classNames, children, errors } = props;

  return (
    <div className={classNames}>
      {children}
      {errors}
    </div>
  );
}

export function DefaultObjectFieldTemplate(props: ObjectFieldTemplateProps) {
  const { properties } = props;
  return <>{properties?.map(p => p.content)}</>;
}
