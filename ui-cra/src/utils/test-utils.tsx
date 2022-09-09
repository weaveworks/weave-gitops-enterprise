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
  ListPipelinesResponse,
  Pipelines,
} from '../api/pipelines/pipelines.pb';
import {
  GetConfigResponse,
  GetPolicyResponse,
  GetPolicyValidationResponse,
  ListGitopsClustersResponse,
  ListPoliciesResponse,
  ListPolicyValidationsResponse,
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

// Give an object that looks like a request so things like .json() work for tests
const mockRes = {
  ok: true,
  clone() {
    return this;
  },
  json() {
    return this;
  },
  then() {},
  catch() {},
};

export const defaultContexts = () => [
  [ThemeProvider, { theme: theme }],
  [MuiThemeProvider, { theme: muiTheme }],
  [
    RequestContextProvider,
    { fetch: () => new Promise(accept => accept(mockRes)) },
  ],
  [QueryClientProvider, { client: new QueryClient() }],
  [
    EnterpriseClientProvider,
    {
      api: new EnterpriseClientMock(),
    },
  ],
  [
    CoreClientContextProvider,
    {
      api: new CoreClientMock(),
    },
  ],
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

  GetFeatureFlags() {
    // FIXME: this is not working
    return promisify({
      flags: {
        WEAVE_GITOPS_FEATURE_CLUSTER: 'true',
      },
    });
  }

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
    this.GetCanary = this.GetCanary.bind(this);
  }
  ListCanariesReturns: ListCanariesResponse = {};
  GetCanaryReturns: GetCanaryResponse = {};
  IsFlaggerAvailableReturns: IsFlaggerAvailableResponse = {};

  ListCanaries() {
    return promisify(this.ListCanariesReturns);
  }

  IsFlaggerAvailable() {
    return promisify(this.IsFlaggerAvailableReturns);
  }

  GetCanary() {
    return promisify(this.GetCanaryReturns);
  }
}

export class PolicyClientMock {
  constructor() {
    this.ListPolicies = this.ListPolicies.bind(this);
    this.ListPolicyValidations = this.ListPolicyValidations.bind(this);
    this.GetPolicy = this.GetPolicy.bind(this);
    this.GetPolicyValidation = this.GetPolicyValidation.bind(this);
  }
  ListPoliciesReturns: ListPoliciesResponse = {};
  ListPolicyValidationsReturns: ListPolicyValidationsResponse = {};
  GetPolicyReturns: GetPolicyResponse = {};
  GetPolicyValidationReturns: GetPolicyValidationResponse = {};

  ListPolicies() {
    return promisify(this.ListPoliciesReturns);
  }
  GetPolicy() {
    return promisify(this.GetPolicyReturns);
  }
  ListPolicyValidations() {
    return promisify(this.ListPolicyValidationsReturns);
  }
  GetPolicyValidation() {
    return promisify(this.GetPolicyValidationReturns);
  }
}

export class PipelinesClientMock implements Pipelines {
  constructor() {
    this.ListPipelines = this.ListPipelines.bind(this);
  }
  ListPipelinesReturns: ListPipelinesResponse = {};
  ListPipelines() {
    return promisify(this.ListPipelinesReturns);
  }
}
export function findCellInCol(cell: string, tableSelector: string) {
  const tbl = document.querySelector(tableSelector);

  const cols = tbl?.querySelectorAll('thead th');
  const idx = findColByHeading(cols, cell) as number;

  const rows = tbl?.querySelectorAll('tbody tr');

  const promotedCell = rows?.item(0).childNodes.item(idx);

  return promotedCell;
}

export function findTextByHeading(
  table: Element,
  row: Element,
  headingName: string,
) {
  const cols = table?.querySelectorAll('thead th');
  const index = findColByHeading(cols, headingName) as number;
  return row.childNodes.item(index).textContent;
}

export function getTableInfo(id: string) {
  const tbl = document.querySelector(`#${id} table`);
  const rows = tbl?.querySelectorAll('tbody tr');
  const headers = tbl?.querySelectorAll('thead tr th');

  return { rows, headers };
}

export function sortTableByColumn(tableId: string, column: string) {
  const btns = document.querySelectorAll<HTMLElement>(
    `#${tableId} table thead tr th button`,
  );
  // Click on ${column} button
  btns.forEach(ele => {
    if (ele.textContent === column) {
      ele.click();
    }
  });
}


// Helper to ensure that tests still pass if columns get re-ordered
function findColByHeading(
  cols: NodeListOf<Element> | undefined,
  heading: string,
): null | number {
  if (!cols) {
    return null;
  }

  let idx = null;
  cols?.forEach((e, i) => {
    //TODO: look for a better matching
    if (e.innerHTML.match('(>' + heading + '<)')) {
      idx = i;
    }
  });

  return idx;
}
