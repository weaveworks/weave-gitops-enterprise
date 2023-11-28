import React from 'react';
import { Terraform } from '../../api/terraform/terraform.pb';
import { ClustersService } from '../../cluster-services/cluster_services.pb';
import {
  UnAuthorizedInterceptor,
  setAPIPathPrefix,
} from '@weaveworks/weave-gitops';
import { ProgressiveDeliveryService } from '@weaveworks/progressive-delivery';
import { Pipelines } from '../../api/pipelines/pipelines.pb';
import { Query } from '../../api/query/query.pb';

interface APIs {
  terraform: typeof Terraform;
  enterprise: typeof ClustersService;
  progressiveDeliveryService: typeof ProgressiveDeliveryService;
  pipelines: typeof Pipelines;
  query: typeof Query;
}

// props
interface Props {
  children: React.ReactNode;
}

export const EnterpriseClientContext = React.createContext<APIs>({} as APIs);

function wrap<T>(api: T): T {
  return UnAuthorizedInterceptor(setAPIPathPrefix(api)) as unknown as T;
}

export const useAPI = () => React.useContext(EnterpriseClientContext);

export function EnterpriseClientProvider({ children }: Props) {
  const api: APIs = {
    terraform: wrap(Terraform),
    enterprise: wrap(ClustersService),
    progressiveDeliveryService: wrap(ProgressiveDeliveryService),
    pipelines: wrap(Pipelines),
    query: wrap(Query),
  };

  return (
    <EnterpriseClientContext.Provider value={api}>
      {children}
    </EnterpriseClientContext.Provider>
  );
}
