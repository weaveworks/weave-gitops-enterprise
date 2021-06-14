export type Param = {
  name: string;
  description?: string;
};

export type Template = {
  name?: string;
  description?: string;
  version?: string;
  parameters?: Param[];
};

export type ListTemplatesResponse = {
  templates?: Template[];
  total?: number;
};
