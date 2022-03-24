import React, { FC, useCallback, useEffect, useState } from 'react';
import { Policy } from '../../types/custom';
import { request } from '../../utils/request';
import { Policies } from './index';
import { useHistory } from 'react-router-dom';
import useNotifications from '../Notifications';

const PoliciesProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(true);
  const [policies, setPolicies] = useState<Policy[]>([]);
  const [policy, setPolicy] = useState<Policy | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const { setNotifications } = useNotifications();

  const history = useHistory();

  const policiesUrl = '/v1/policies';

  // const getPolicy = (policyName: string) =>
  //   policies.find(policy => policy.name === policyName) || null;

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

  const getPolicy = useCallback((policyName) => {
    setLoading(true);
    request('GET', `${policiesUrl}/${policyName}`, {
      cache: 'no-store',
    })
      .then(res => setPolicies(res.policies))
      .catch(err =>
        setNotifications([{ message: err.message, variant: 'danger' }]),
      )
      .finally(() => setLoading(false));
  }, [setNotifications]);


  const getPolicies = useCallback(() => {
    setLoading(true);
    request('GET', policiesUrl, {
      cache: 'no-store',
    })
      .then(res => setPolicies(res.policies))
      .catch(err =>
        setNotifications([{ message: err.message, variant: 'danger' }]),
      )
      .finally(() => setLoading(false));
  }, [setNotifications]);

  useEffect(() => {
    getPolicies();
    return history.listen(getPolicies);
  }, [history, getPolicies]);

  return (
    <Policies.Provider
      value={{
        policies,
        policy,
        loading,
        error,
        setError,
        getPolicy,
      }}
    >
      {children}
    </Policies.Provider>
  );
};

export default PoliciesProvider;
