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
  annotations?: string[];
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

export type Maintainer = {
  name?: string;
  email?: string;
  url?: string;
};

export type HelmRepository = {
  name?: string;
  namespace?: string;
};

export type Profile = {
  name: string;
  home?: string;
  sources?: string[];
  description?: string;
  keywords?: string[];
  maintainers?: Maintainer[];
  icon?: string;
  annotations?: { [key: string]: string };
  kubeVersion?: string;
  helmRepository?: HelmRepository;
  availableVersions: string[];
  layer?: string;
};

export type ListProfilesResponse = {
  profiles?: Profile[];
};

export type UpdatedProfile = {
  name: Profile['name'];
  values: { version: string; yaml: string; selected?: boolean }[];
  required: boolean;
  layer?: string;
};

export type ChildrenOccurences = {
  name: string;
  groupVisible: boolean;
  count: number;
};
