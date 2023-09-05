import { coreClient } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import { Terraform } from '../api/terraform/terraform.pb';
import { RequestMethod } from '../types/global';

export const processResponse = (res: Response) => {
  // 400s / 500s have res.ok = false
  if (!res.ok) {
    return res
      .clone()
      .json()
      .catch(() => res.text().then(message => ({ message })))
      .then(data => Promise.reject(data));
  }
  return res
    .clone()
    .json()
    .catch(() => res.text().then(message => ({ success: true, message })));
};

export const processEntitlementHeaders = (res: Response) => {
  return res.headers?.get('Entitlement-Expired-Message');
};

export const rawRequest = (
  method: RequestMethod,
  query: RequestInfo,
  options: RequestInit = {},
  fetchArg?: typeof window.fetch,
) => {
  const f = fetchArg || window.fetch;
  return f(query, { ...options, method }).then(res => processResponse(res));
};

export const requestWithEntitlementHeader = (
  method: RequestMethod,
  query: RequestInfo,
  options: RequestInit = {},
  fetchArg?: typeof window.fetch,
) => {
  const f = fetchArg || window.fetch;
  return f(query, { ...options, method }).then(res =>
    processResponse(res).then(body => ({
      data: body,
      entitlement: processEntitlementHeaders(res),
    })),
  );
};

export const removeToken = (provider: string) =>
  localStorage.removeItem(`gitProviderToken_${provider}`);

//mutates the OSS api to get Terraform Sync/Suspend down into the CheckboxActions component without modifying OSS
export function addTFSupport(client: typeof coreClient) {
  const originalSync = client.SyncFluxObject;
  client.SyncFluxObject = function (req, initReq) {
    const objects = req.objects || [];
    const isTerraform = _.every(objects, obj => obj.kind === 'Terraform');
    if (isTerraform) {
      return Terraform.SyncTerraformObjects({ objects: objects }, initReq);
    }

    return originalSync.call(
      this,
      {
        objects: objects,
        withSource: req.withSource,
      },
      initReq,
    );
  };

  const originalSuspend = client.ToggleSuspendResource;
  client.ToggleSuspendResource = function (req, initReq) {
    const objects = req.objects || [];
    const isTerraform = _.every(objects, obj => obj.kind === 'Terraform');
    if (isTerraform) {
      return Terraform.ToggleSuspendTerraformObjects(
        { objects: objects, suspend: req.suspend },
        initReq,
      );
    }

    return originalSuspend.call(
      this,
      {
        objects: objects,
        suspend: req.suspend,
      },
      initReq,
    );
  };
}
