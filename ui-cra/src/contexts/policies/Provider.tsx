import React, { FC, useCallback, useEffect, useState } from 'react';
import { Template } from '../../types/custom';
import { request } from '../../utils/request';
import { Policies } from './index';
import { useHistory } from 'react-router-dom';
import useNotifications from '../Notifications';

const PoliciesProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(true);
  const [policies, setPolicies] = useState<Template[]>([]);
  const [activePolicy, setActivePolicy] = useState<Template | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const { setNotifications } = useNotifications();

  // const history = useHistory();

  // const templatesUrl = '/v1/templates';

  const getPolicy = (policyName: string) =>
    policies.find(policy => policy.name === policyName) || null;

  // const renderTemplate = useCallback(
  //   data => {
  //     setLoading(true);
  //     request('POST', `${templatesUrl}/${activeTemplate?.name}/render`, {
  //       body: JSON.stringify(data),
  //     })
  //       .then(data => setPRPreview(data.renderedTemplate))
  //       .catch(err =>
  //         setNotifications([{ message: err.message, variant: 'danger' }]),
  //       )
  //       .finally(() => setLoading(false));
  //   },
  //   [activeTemplate, setNotifications],
  // );

  // const addCluster = useCallback(({ ...data }, token: string) => {
  //   setLoading(true);
  //   return request('POST', '/v1/clusters', {
  //     body: JSON.stringify(data),
  //     headers: new Headers({ 'Git-Provider-Token': `token ${token}` }),
  //   }).finally(() => setLoading(false));
  // }, []);

  // const getTemplates = useCallback(() => {
  //   setLoading(true);
  //   request('GET', templatesUrl, {
  //     cache: 'no-store',
  //   })
  //     .then(res => setTemplates(res.templates))
  //     .catch(err =>
  //       setNotifications([{ message: err.message, variant: 'danger' }]),
  //     )
  //     .finally(() => setLoading(false));
  // }, [setNotifications]);

  // useEffect(() => {
  //   getTemplates();
  //   return history.listen(getTemplates);
  // }, [history, getTemplates]);

  return (
    <Policies.Provider
      value={{
        policies,
        loading,
        activePolicy,
        setActivePolicy,
        getPolicy,
        error,
        setError,
      }}
    >
      {children}
    </Policies.Provider>
  );
};

export default PoliciesProvider;
