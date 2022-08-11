import { createContext, Dispatch, useContext } from 'react';
import { Template } from '../../cluster-services/cluster_services.pb';

interface TemplatesContext {
  templates: Template[] | undefined;
  loading: boolean;
  activeTemplate: Template | null;
  setActiveTemplate: Dispatch<React.SetStateAction<Template | null>>;
  addCluster: (data: any, token: string, templateKind: string) => Promise<any>;
  renderTemplate: (data: any) => Promise<any>;
  getTemplate: (templateName: string) => Template | null;
  isLoading: boolean;
}

export const Templates = createContext<TemplatesContext | null>(null);

export default () => useContext(Templates) as TemplatesContext;
