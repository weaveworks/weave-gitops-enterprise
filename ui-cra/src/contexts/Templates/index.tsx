import { createContext, Dispatch, useContext } from 'react';
import { Template } from '../../cluster-services/cluster_services.pb';

interface TemplatesContext {
  templates: Template[] | undefined;
  loading: boolean;
  activeTemplate: Template | null;
  setActiveTemplate: Dispatch<React.SetStateAction<Template | null>>;
  error: string | null;
  addCluster: (data: any, token: string) => Promise<any>;
  renderTemplate: (data: any) => void;
  getTemplate: (templateName: string) => Template | null;
  PRPreview: string | null;
  setPRPreview: Dispatch<React.SetStateAction<string | null>>;
  setError: Dispatch<React.SetStateAction<string | null>>;
}

export const Templates = createContext<TemplatesContext | null>(null);

export default () => useContext(Templates) as TemplatesContext;
