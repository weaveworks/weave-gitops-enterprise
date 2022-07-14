import { MuiThemeProvider } from '@material-ui/core';
import {
  GetCanaryResponse,
  IsFlaggerAvailableResponse,
  ListCanariesResponse,
  ProgressiveDeliveryService,
} from '@weaveworks/progressive-delivery';
import { CoreClientContextProvider, theme } from '@weaveworks/weave-gitops';
import {
  ListHelmReleasesResponse,
  ListKustomizationsResponse,
} from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';
import _ from 'lodash';
import React from 'react';
import { QueryClient, QueryClientProvider } from 'react-query';
import { MemoryRouter } from 'react-router-dom';
import { ThemeProvider } from 'styled-components';
import {
  GetConfigResponse,
  ListGitopsClustersResponse,
  ListTemplatesResponse,
} from '../cluster-services/cluster_services.pb';
import Compose from '../components/ProvidersCompose';
import ClustersProvider from '../contexts/Clusters/Provider';
import EnterpriseClientProvider from '../contexts/EnterpriseClient/Provider';
import NotificationProvider from '../contexts/Notifications/Provider';
import RequestContextProvider from '../contexts/Request';
import TemplatesProvider from '../contexts/Templates/Provider';
import { muiTheme } from '../muiTheme';

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

export const defaultContexts = () => [
  [ThemeProvider, { theme: theme }],
  [MuiThemeProvider, { theme: muiTheme }],
  [
    RequestContextProvider,
    { fetch: () => new Promise(accept => accept(null)) },
  ],
  [QueryClientProvider, { client: new QueryClient() }],
  [
    EnterpriseClientProvider,
    {
      api: new EnterpriseClientMock(),
    },
  ],
  [CoreClientContextProvider, { api: new CoreClientMock() }],
  [MemoryRouter],
  [NotificationProvider],
  [TemplatesProvider],
  [ClustersProvider],
];

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

export class ProgressiveDeliveryMock implements ProgressiveDeliveryService {
  constructor() {
    this.ListCanaries = this.ListCanaries.bind(this);
    this.GetCanary = this.GetCanary.bind(this);
    this.IsFlaggerAvailable = this.IsFlaggerAvailable.bind(this);
  }
  ListCanariesReturns: ListCanariesResponse = {};
  GetCanaryReturns: GetCanaryResponse = {};
  IsFlaggerAvailableReturns: IsFlaggerAvailableResponse = {};

  ListCanaries() {
    return promisify(this.ListCanariesReturns);
  }

  GetCanary() {
    return promisify(this.GetCanaryReturns);
  }

  IsFlaggerAvailable() {
    return promisify(this.IsFlaggerAvailableReturns);
  }
}
