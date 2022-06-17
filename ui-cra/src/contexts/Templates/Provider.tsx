import React, { FC, useCallback, useContext, useState } from 'react';
import { useQuery } from 'react-query';
import { request } from '../../utils/request';
import { Templates } from './index';
import useNotifications from './../Notifications';
import { EnterpriseClientContext } from '../EnterpriseClient';
import {
  ListTemplatesResponse,
  Template,
} from '../../cluster-services/cluster_services.pb';

const TemplatesProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(true);
  const [templates, setTemplates] = useState<Template[] | undefined>([]);
  const [activeTemplate, setActiveTemplate] = useState<Template | null>(null);
  const [PRPreview, setPRPreview] = useState<string | null>(null);
  const { notifications, setNotifications } = useNotifications();
  const { api } = useContext(EnterpriseClientContext);

  const templatesUrl = '/v1/templates';

  const getTemplate = (templateName: string) =>
    templates?.find(template => template.name === templateName) || null;

  const renderTemplate = useCallback(
    data => {
      setLoading(true);
      request('POST', `${templatesUrl}/${activeTemplate?.name}/render`, {
        body: JSON.stringify(data),
      })
        .then(data => setPRPreview(data.renderedTemplate))
        .catch(err =>
          setNotifications([
            { message: { text: err.message }, variant: 'danger' },
          ]),
        )
        .finally(() => setLoading(false));
    },
    [activeTemplate, setNotifications],
  );

  const addCluster = useCallback(({ ...data }, token: string) => {
    setLoading(true);
    return request('POST', '/v1/clusters', {
      body: JSON.stringify(data),
      headers: new Headers({ 'Git-Provider-Token': `token ${token}` }),
    }).finally(() => setLoading(false));
  }, []);

  const onError = (error: Error) => {
    if (
      error &&
      notifications?.some(
        notification => error.message === notification.message.text,
      ) === false
    ) {
      setNotifications([
        ...notifications,
        { message: { text: error.message }, variant: 'danger' },
      ]);
    }
  };

  const onSuccess = (data: ListTemplatesResponse) =>
    setTemplates(data.templates);

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
        activeTemplate,
        setActiveTemplate,
        getTemplate,
        addCluster,
        renderTemplate,
        PRPreview,
        setPRPreview,
      }}
    >
      {children}
    </Templates.Provider>
  );
};

export default TemplatesProvider;
