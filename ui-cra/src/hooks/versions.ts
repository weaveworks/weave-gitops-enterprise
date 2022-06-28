import { useContext } from 'react';
import { useQuery } from 'react-query';
import { requestWithEntitlementHeader } from '../utils/request';
import { EnterpriseClientContext } from '../contexts/EnterpriseClient';
import { GetConfigResponse } from '../cluster-services/cluster_services.pb';

export function useListVersion() {
  return useQuery<{ data: any; entitlement: string | null }, Error>(
    'version',
    () => requestWithEntitlementHeader('GET', '/v1/enterprise/version'),
  );
}

export function useListConfig() {
  const { api } = useContext(EnterpriseClientContext);
  return useQuery<GetConfigResponse, Error>('config', () => api.GetConfig({}));
}
