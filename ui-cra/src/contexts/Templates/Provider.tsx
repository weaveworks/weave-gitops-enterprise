import React, { FC, useCallback, useEffect, useState } from 'react';
import { Template } from '../../types/custom';
import { request } from '../../utils/request';
import { Templates } from './index';
import { useHistory } from 'react-router-dom';
import useNotifications from './../Notifications';

const TemplatesProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(true);
  const [templates, setTemplates] = useState<Template[]>([]);
  const [activeTemplate, setActiveTemplate] = useState<Template | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const [PRPreview, setPRPreview] = useState<string | null>(null);
  const [creatingPR, setCreatingPR] = useState<boolean>(false);
  const { setNotifications } = useNotifications();

  const history = useHistory();

  const templatesUrl = '/v1/templates';

  const getTemplate = (templateName: string) =>
    templates.find(template => template.name === templateName) || null;

  const renderTemplate = useCallback(
    data => {
      setLoading(true);
      request('POST', `${templatesUrl}/${activeTemplate?.name}/render`, {
        body: JSON.stringify(data),
      })
        .then(data => setPRPreview(data.renderedTemplate))
        .catch(err =>
          setNotifications([{ message: err.message, variant: 'danger' }]),
        )
        .finally(() => setLoading(false));
    },
    [activeTemplate, setNotifications],
  );

  const addCluster = useCallback(
    ({ ...data }) => {
      setCreatingPR(true);
      request('POST', '/v1/clusters', {
        body: JSON.stringify(data),
      })
        .then(() => history.push('/clusters'))
        .catch(err =>
          setNotifications([{ message: err.message, variant: 'danger' }]),
        )
        .finally(() => setCreatingPR(false));
    },
    [history, setNotifications],
  );

  const getTemplates = useCallback(() => {
    setLoading(true);
    request('GET', templatesUrl, {
      cache: 'no-store',
    })
      .then(res => setTemplates(res.templates))
      .catch(err =>
        setNotifications([{ message: err.message, variant: 'danger' }]),
      )
      .finally(() => setLoading(false));
  }, [setNotifications]);

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
        creatingPR,
      }}
    >
      {children}
    </Templates.Provider>
  );
};

export default TemplatesProvider;
