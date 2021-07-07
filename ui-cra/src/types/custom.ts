import { JSONSchema7 } from 'json-schema';
import { ReactElement } from 'react';
import { IdSchema, UiSchema } from 'react-jsonschema-form';

export type Param = {
  name: string;
  description?: string;
  options?: string[];
};

export type Object = {
  kind: string;
  apiVersion: string;
  parameters: Param['name'];
};

export type Template = {
  name?: string;
  description?: string;
  version?: string;
  parameters?: Param[];
  objects: Object[];
};

export type ListTemplatesResponse = {
  templates?: Template[];
  total?: number;
};
