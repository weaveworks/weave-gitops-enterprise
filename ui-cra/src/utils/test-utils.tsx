import {
  ListHelmReleasesResponse,
  ListKustomizationsResponse,
} from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';
import _ from 'lodash';
import React from 'react';
import {
  GetConfigResponse,
  ListGitopsClustersResponse,
  ListTemplatesResponse,
} from '../cluster-services/cluster_services.pb';
import Compose from '../components/ProvidersCompose';

export const withContext = (contexts: any[]) => {
  return (component: React.ReactElement) => {
    const tree = _.reduce(
      contexts,
      (r: any[], c) => {
        const [Ctx, props] = c;

        r.push((otherProps: any) => <Ctx {...props} {...otherProps} />);
        return r;
      },
      [],
    );

    return <Compose components={tree}>{component}</Compose>;
  };
};

const promisify = <R, E>(res: R, errRes?: E) =>
  new Promise<R>((accept, reject) => {
    if (errRes) {
      return reject(errRes);
    }
    accept(res);
  });

export class EnterpriseClientMock {
  constructor() {
    this.GetConfig = this.GetConfig.bind(this);
    this.ListTemplates = this.ListTemplates.bind(this);
    this.ListGitopsClusters = this.ListGitopsClusters.bind(this);
  }
  GetConfigReturns: GetConfigResponse = {};
  ListTemplatesReturns: ListTemplatesResponse = {};
  ListGitopsClustersResponse: ListGitopsClustersResponse = {};

  GetConfig() {
    return promisify(this.GetConfigReturns);
  }

  ListTemplates() {
    return promisify(this.ListTemplatesReturns);
  }

  ListGitopsClusters() {
    return promisify(this.ListGitopsClustersResponse);
  }
}

export class CoreClientMock {
  constructor() {
    this.ListKustomizations = this.ListKustomizations.bind(this);
    this.ListHelmReleases = this.ListHelmReleases.bind(this);
  }
  ListKustomizationsReturns: ListKustomizationsResponse = {};
  ListHelmReleasesReturns: ListHelmReleasesResponse = {};

  ListKustomizations() {
    return promisify(this.ListKustomizationsReturns);
  }

  ListHelmReleases() {
    return promisify(this.ListHelmReleasesReturns);
  }
}
