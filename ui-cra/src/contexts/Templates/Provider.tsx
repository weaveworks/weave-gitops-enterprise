import { FC, useCallback, useContext, useState } from 'react';
import { useQuery } from 'react-query';
import { ListTemplatesResponse } from '../../cluster-services/cluster_services.pb';
import { TemplateEnriched } from '../../types/custom';
import { request } from '../../utils/request';
import { EnterpriseClientContext } from '../EnterpriseClient';
import useNotifications from './../Notifications';
import { Templates } from './index';

const TemplatesProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(false);
  const [templates, setTemplates] = useState<TemplateEnriched[] | undefined>(
    [],
  );
  const { setNotifications } = useNotifications();
  const { api } = useContext(EnterpriseClientContext);

  const templatesUrl = '/v1/templates';

  const getTemplate = (templateName: string) =>
    templates?.find(template => template.name === templateName) || null;

  const renderTemplate = useCallback((templateName, data) => {
    return request('POST', `${templatesUrl}/${templateName}/render`, {
      body: JSON.stringify(data),
    });
  }, []);

  const renderKustomization = (data: any) => {
    return request('POST', `kustomization/render`, {
      body: JSON.stringify(data),
    });
  };

  const addCluster = useCallback(
    ({ ...data }, token: string, templateKind: string) => {
      setLoading(true);
      return request(
        'POST',
        templateKind === 'GitOpsTemplate'
          ? '/v1/tfcontrollers'
          : '/v1/clusters',
        {
          body: JSON.stringify(data),
          headers: new Headers({ 'Git-Provider-Token': `token ${token}` }),
        },
      ).finally(() => setLoading(false));
    },
    [],
  );

  const onError = (error: Error) =>
    setNotifications([{ message: { text: error.message }, variant: 'danger' }]);

  const onSuccess = (data: ListTemplatesResponse) =>
    setTemplates(data.templates as TemplateEnriched[]);

  const { isLoading } = useQuery<ListTemplatesResponse, Error>(
    'templates',
    () => api.ListTemplates({}),
    {
      keepPreviousData: true,
      onSuccess,
      onError,
    },
  );

  return (
    <Templates.Provider
      value={{
        isLoading,
        templates,
        loading,
        getTemplate,
        addCluster,
        renderTemplate,
        renderKustomization,
      }}
    >
      {children}
    </Templates.Provider>
  );
};

export default TemplatesProvider;
