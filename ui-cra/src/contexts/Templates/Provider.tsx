import React, { FC, useCallback, useContext, useEffect, useState } from 'react';
import { request } from '../../utils/request';
import { Templates } from './index';
import { useHistory } from 'react-router-dom';
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
  const [error, setError] = React.useState<string | null>(null);
  const [PRPreview, setPRPreview] = useState<string | null>(null);
  const { setNotifications } = useNotifications();
  const { api } = useContext(EnterpriseClientContext);

  const history = useHistory();

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

  const getTemplates = useCallback(() => {
    setLoading(true);
    api
      .ListTemplates({})
      .then((res: ListTemplatesResponse) => setTemplates(res.templates))
      .catch((err: Error) =>
        setNotifications([
          { message: { text: err.message }, variant: 'danger' },
        ]),
      )
      .finally(() => setLoading(false));
  }, [api, setNotifications]);

  useEffect(() => {
    getTemplates();
    return history.listen(getTemplates);
  }, [history, getTemplates]);

  return (
    <Templates.Provider
      value={{
        templates,
        loading,
        activeTemplate,
        setActiveTemplate,
        getTemplate,
        error,
        setError,
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
