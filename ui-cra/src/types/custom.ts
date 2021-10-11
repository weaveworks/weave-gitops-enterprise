export type Param = {
  name: string;
  description?: string;
  options?: string[];
};

export type TemplateObject = {
  kind?: string;
  apiVersion?: string;
  parameters?: Param['name'];
  name?: string;
  displayName?: string;
};

export type Template = {
  name?: string;
  description?: string;
  version?: string;
  parameters?: Param[];
  objects?: TemplateObject[];
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

export enum GitProviderName {
  GitHub = 'github',
  Gitlab = 'gitlab',
}
