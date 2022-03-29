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

  const getPolicy = useCallback((policyName) => {
    setLoading(true);
    request('GET', `${policiesUrl}/${policyName}`, {
      cache: 'no-store',
    })
      .then(res => setPolicy(res.policy))
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
    return () => {
      setPolicies([])
    };
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
        setPolicy,
      }}
    >
      {children}
    </Policies.Provider>
  );
};

export default PoliciesProvider;
