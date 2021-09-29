import { JSONSchema7 } from 'json-schema';
import { ReactElement } from 'react';
import { IdSchema, UiSchema } from 'react-jsonschema-form';

export type Param = {
  name: string;
  description?: string;
  options?: string[];
};

export type Object = {
  kind?: string;
  apiVersion?: string;
  parameters?: Param['name'];
  name?: string;
};

export type Template = {
  name?: string;
  description?: string;
  version?: string;
  parameters?: Param[];
  objects?: Object[];
  error?: string;
  provider?: string;
};

export type ListTemplatesResponse = {
  templates?: Template[];
  total?: number;
};

export type Credential = {
  group?: string;
  version?: string;
  kind?: string;
  name?: string;
  namespace?: string;
};

export type ListCredentialsResponse = {
  credentials?: Credential[];
  total?: number;
};
