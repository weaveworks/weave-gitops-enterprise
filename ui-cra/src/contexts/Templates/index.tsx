import { createContext, useContext } from 'react';
import { TemplateEnriched } from '../../types/custom';

interface TemplatesContext {
  templates: TemplateEnriched[] | undefined;
  loading: boolean;
  addCluster: (data: any, token: string, templateKind: string) => Promise<any>;
  renderTemplate: (templateName: string, data: any) => Promise<any>;
  renderKustomization: (data: any) => Promise<any>;
  getTemplate: (templateName: string) => TemplateEnriched | null;
  isLoading: boolean;
}

export const Templates = createContext<TemplatesContext | null>(null);

export default () => useContext(Templates) as TemplatesContext;
