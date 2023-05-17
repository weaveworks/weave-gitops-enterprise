import { MuiThemeProvider } from '@material-ui/core';
import {
  GetCanaryResponse,
  IsFlaggerAvailableResponse,
  ListCanariesResponse,
  ProgressiveDeliveryService,
} from '@weaveworks/progressive-delivery';
import { CoreClientContextProvider, theme } from '@weaveworks/weave-gitops';
import {
  GetObjectRequest,
  GetObjectResponse,
  IsCRDAvailableRequest,
  IsCRDAvailableResponse,
  ListObjectsRequest,
  ListObjectsResponse,
} from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';
import _ from 'lodash';
import React from 'react';
import { QueryCache, QueryClient, QueryClientProvider } from 'react-query';
import { MemoryRouter } from 'react-router-dom';
import { ThemeProvider } from 'styled-components';
import {
  GetGithubAuthStatusResponse,
  GetGithubDeviceCodeResponse,
  GetGitlabAuthURLResponse,
  ParseRepoURLResponse,
  ValidateProviderTokenResponse,
} from '../api/gitauth/gitauth.pb';
import {
  GetPipelineResponse,
  ListPipelinesResponse,
  Pipelines,
} from '../api/pipelines/pipelines.pb';
import {
  GetTerraformObjectPlanResponse,
  GetTerraformObjectResponse,
  ListTerraformObjectsResponse,
  ReplanTerraformObjectResponse,
  SyncTerraformObjectResponse,
  Terraform,
  ToggleSuspendTerraformObjectResponse,
} from '../api/terraform/terraform.pb';
import {
  GetConfigResponse,
  GetExternalSecretResponse,
  GetPolicyConfigResponse,
  GetPolicyResponse,
  GetPolicyValidationResponse,
  GetWorkspacePoliciesResponse,
  GetWorkspaceResponse,
  GetWorkspaceRoleBindingsResponse,
  GetWorkspaceRolesResponse,
  GetWorkspaceServiceAccountsResponse,
  ListEventsResponse,
  ListExternalSecretsResponse,
  ListGitopsClustersResponse,
  ListPoliciesResponse,
  ListPolicyConfigsResponse,
  ListPolicyValidationsResponse,
  ListTemplatesResponse,
  ListWorkspacesResponse,
} from '../cluster-services/cluster_services.pb';
import Compose from '../components/ProvidersCompose';
import EnterpriseClientProvider from '../contexts/EnterpriseClient/Provider';
import { GitAuthProvider } from '../contexts/GitAuth';
import NotificationProvider from '../contexts/Notifications/Provider';
import RequestContextProvider from '../contexts/Request';
import { muiTheme } from '../muiTheme';

export type RequestError = Error & {
  code?: number;
};

export function withTheme(element: any) {
  return (
    <MuiThemeProvider theme={muiTheme}>
      <ThemeProvider theme={theme}>{element}</ThemeProvider>
    </MuiThemeProvider>
  );
}

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
  [
    QueryClientProvider,
    {
      client: new QueryClient({
        queryCache: new QueryCache({
          onError: error => {
            const err = error as { code: number; message: string };
            const { pathname, search } = window.location;
            const redirectUrl = encodeURIComponent(`${pathname}${search}`);

            if (err.code === 401) {
              window.location.href = `/sign_in?redirect=${redirectUrl}`;
            }
          },
        }),
      }),
    },
  ],
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
  [GitAuthProvider, { api: new ApplicationsClientMock() }],
];

export const promisify = <R, E>(res: R, errRes?: E) =>
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

const defaultListObjectsResponse: ListObjectsResponse = {
  objects: [],
  errors: [],
};

export class CoreClientMock {
  constructor() {
    this.ListObjects = this.ListObjects.bind(this);
    this.GetFeatureFlags = this.GetFeatureFlags.bind(this);
    this.GetObject = this.GetObject.bind(this);
    this.IsCRDAvailable = this.IsCRDAvailable.bind(this);
  }
  GetFeatureFlagsReturns: { flags: { [x: string]: string } } = {
    flags: {
      WEAVE_GITOPS_FEATURE_CLUSTER: 'true',
      WEAVE_GITOPS_FEATURE_TENANCY: 'true',
    },
  };
  ListObjectsReturns: { [kind: string]: ListObjectsResponse } = {};
  GetObjectReturns: GetObjectResponse = {};
  IsCRDAvailableReturn: { [name: string]: IsCRDAvailableResponse } = {};

  GetFeatureFlags() {
    return promisify(this.GetFeatureFlagsReturns);
  }

  ListObjects(req: ListObjectsRequest) {
    return promisify(
      this.ListObjectsReturns[req.kind!] || defaultListObjectsResponse,
    );
  }

  GetObject(req: GetObjectRequest) {
    return promisify(this.GetObjectReturns);
  }
  IsCRDAvailable(req: IsCRDAvailableRequest) {
    return promisify(this.IsCRDAvailableReturn[req.name!]);
  }
}

export class ApplicationsClientMock {
  constructor() {
    this.GetGithubDeviceCode = this.GetGithubDeviceCode.bind(this);
    this.GetGithubAuthStatus = this.GetGithubAuthStatus.bind(this);
    this.ParseRepoURL = this.ParseRepoURL.bind(this);
    this.GetGitlabAuthURL = this.GetGitlabAuthURL.bind(this);
    this.ValidateProviderToken = this.ValidateProviderToken.bind(this);
  }
  GetGithubDeviceCodeReturn: GetGithubDeviceCodeResponse = {};
  GetGithubAuthStatusReturn: GetGithubAuthStatusResponse = {};
  ParseRepoURLReturn: ParseRepoURLResponse = {};
  GetGitlabAuthURLReturn: GetGitlabAuthURLResponse = {};
  ValidateProviderTokenReturn: ValidateProviderTokenResponse = {};

  GetGithubDeviceCode() {
    return promisify(this.GetGithubDeviceCodeReturn);
  }

  GetGithubAuthStatus() {
    return promisify(this.GetGithubAuthStatusReturn);
  }

  ParseRepoURL() {
    return promisify(this.ParseRepoURLReturn);
  }

  GetGitlabAuthURL(req: any) {
    return promisify(this.GetGitlabAuthURLReturn);
  }

  ValidateProviderToken() {
    return promisify(this.ValidateProviderTokenReturn);
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

export class PolicyConfigsClientMock {
  ListPolicyConfigsReturns: ListPolicyConfigsResponse = {};
  GetPolicyConfigReturns: GetPolicyConfigResponse = {};
  
  ListPolicyConfigs() {
    return promisify(this.ListPolicyConfigsReturns);
  }

  GetPolicyConfig() {
    return promisify(this.GetPolicyConfigReturns);
  }
}

export class PipelinesClientMock implements Pipelines {
  constructor() {
    this.ListPipelines = this.ListPipelines.bind(this);
    this.GetPipeline = this.GetPipeline.bind(this);
  }
  ListPipelinesReturns: ListPipelinesResponse = {};
  GetPipelineReturns: GetPipelineResponse = {};
  ErrorRef: { code: number; message: string } | undefined;

  ListPipelines() {
    return promisify(this.ListPipelinesReturns, this.ErrorRef);
  }

  GetPipeline() {
    return promisify(this.GetPipelineReturns);
  }
}

export class TerraformClientMock implements Terraform {
  constructor() {
    this.ListTerraformObjects = this.ListTerraformObjects.bind(this);
    this.GetTerraformObject = this.GetTerraformObject.bind(this);
    this.SyncTerraformObject = this.SyncTerraformObject.bind(this);
    this.ToggleSuspendTerraformObject =
      this.ToggleSuspendTerraformObject.bind(this);
  }

  ListTerraformObjectsReturns: ListTerraformObjectsResponse = {};
  ListTerraformObjects() {
    return promisify(this.ListTerraformObjectsReturns);
  }

  GetTerraformObjectReturns: GetTerraformObjectResponse = {};
  GetTerraformObject() {
    return promisify(this.GetTerraformObjectReturns);
  }

  GetTerraformObjectPlanReturns: GetTerraformObjectPlanResponse = {
    enablePlanViewing: true,
  };
  GetTerraformObjectPlan() {
    return promisify(this.GetTerraformObjectPlanReturns);
  }

  SyncTerraformObjectReturns: SyncTerraformObjectResponse = {};
  SyncTerraformObject() {
    return promisify(this.SyncTerraformObjectReturns);
  }

  ToggleSuspendTerraformObjectReturns: ToggleSuspendTerraformObjectResponse =
    {};
  ToggleSuspendTerraformObject() {
    return promisify(this.ToggleSuspendTerraformObjectReturns);
  }

  ReplanTerraformObjectReturns: ReplanTerraformObjectResponse = {};
  ReplanTerraformObject() {
    return promisify(this.ReplanTerraformObjectReturns);
  }
}

export class WorkspaceClientMock {
  ListWorkspacesReturns: ListWorkspacesResponse = {};
  GetWorkspaceReturns: GetWorkspaceResponse = {};
  GetWorkspaceServiceAccountsReturns: GetWorkspaceServiceAccountsResponse = {};
  GetWorkspaceRolesReturns: GetWorkspaceRolesResponse = {};
  GetWorkspaceRoleBindingsReturns: GetWorkspaceRoleBindingsResponse = {};
  GetWorkspacePoliciesReturn: GetWorkspacePoliciesResponse = {};
  ListWorkspaces = () => promisify(this.ListWorkspacesReturns);

  GetWorkspace = () => promisify(this.GetWorkspaceReturns);

  GetWorkspaceServiceAccounts = () =>
    promisify(this.GetWorkspaceServiceAccountsReturns);

  GetWorkspaceRoles = () => promisify(this.GetWorkspaceRolesReturns);
  GetWorkspaceRoleBindings = () =>
    promisify(this.GetWorkspaceRoleBindingsReturns);
  GetWorkspacePolicies = () => promisify(this.GetWorkspacePoliciesReturn);
}

export class SecretsClientMock {
  ListSecretsReturns: ListExternalSecretsResponse = {};
  GetExternalSecretReturns: GetExternalSecretResponse = {};
  ListEventsReturns: ListEventsResponse = {};

  ListExternalSecrets() {
    return promisify(this.ListSecretsReturns);
  }

  GetExternalSecret() {
    return promisify(this.GetExternalSecretReturns);
  }
  ListEvents() {
    return promisify(this.ListEventsReturns);
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
export function getRowInfoByIndex(tableId: string, rowIndex: number) {
  const rows = document.querySelectorAll(`#${tableId} tbody tr`);
  return rows[rowIndex].querySelectorAll('td');
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

// WIP - Make a sharable class to test all Filterable table functionality

export class TestFilterableTable {
  constructor(_tableId: string, _fireEvent: any) {
    this.tableId = _tableId;
    this.fireEvent = _fireEvent;
  }
  tableId: string = '';
  fireEvent: any;

  getTableInfo() {
    const tbl = document.querySelector(`#${this.tableId} table`);
    const rows = tbl?.querySelectorAll('tbody tr');
    const headers = tbl?.querySelectorAll('thead tr th');
    return { rows, headers };
  }
  getRowInfoByIndex(rowIndex: number) {
    const rows = document.querySelectorAll(`#${this.tableId} tbody tr`);
    return rows[rowIndex].querySelectorAll('td');
  }

  sortTableByColumn(columnName: string) {
    const btns = document.querySelectorAll<HTMLElement>(
      `#${this.tableId} table thead tr th button`,
    );
    btns.forEach(ele => {
      if (ele.textContent === columnName) {
        ele.click();
      }
    });
  }
  searchTableByValue(searchVal: string) {
    const searchBtn = document.querySelector<HTMLElement>(
      `#${this.tableId} button[class*='SearchField']`,
    );
    searchBtn?.click();
    const searchInput = document.getElementById(
      'table-search',
    ) as HTMLInputElement;

    this.fireEvent.change(searchInput, { target: { value: searchVal } });

    const searchForm = document.querySelector(
      `#${this.tableId} div[class*='SearchField'] > form`,
    ) as Element;

    this.fireEvent.submit(searchForm);
    return this.getTableInfo();
  }
  clearSearchByVal(searchVal: string) {
    document.querySelectorAll('.filter-options-chip').forEach(chip => {
      if (chip.textContent === searchVal) {
        const deleteIcon = chip.querySelector('.MuiChip-deleteIcon');
        this.fireEvent.click(deleteIcon);
      }
    });
    const chip = Array.from(
      document.querySelectorAll('.filter-options-chip'),
    ).filter(ele => ele.textContent === searchVal);
    expect(chip.length).toEqual(0);
  }
  applyFilterByValue(filterIndex: number, value: string) {
    const filterBtn = document.querySelector<HTMLElement>(
      `#${this.tableId} button[class*='FilterableTable']`,
    );
    filterBtn?.click();

    const filters = document.querySelectorAll<HTMLElement>(
      `#${this.tableId} form > ul > li`,
    );
    const filterInput = filters[filterIndex].querySelector<HTMLElement>(
      `input[id="${value}"]`,
    );
    filterInput?.click();
    return this.getTableInfo();
  }

  testRowValues = (rowValue: NodeListOf<Element>, matches: Array<string>) => {
    for (let index = 0; index < rowValue.length; index++) {
      const element = rowValue[index];
      expect(element.textContent).toEqual(matches[index]);
    }
  };

  testRenderTable(displayedHeaders: Array<string>, rowLength: number) {
    const { rows, headers } = this.getTableInfo();
    expect(headers).toHaveLength(displayedHeaders.length);
    expect(rows).toHaveLength(rowLength);
    this.testRowValues(headers!, displayedHeaders);
  }

  testSearchTableByValue(searchValue: string, rowValues: Array<Array<string>>) {
    const { rows } = this.searchTableByValue(searchValue);
    expect(rows).toHaveLength(rowValues.length);
    rowValues.forEach((row, index) => {
      const tds = rows![index].querySelectorAll('td');
      this.testRowValues(tds, row);
    });
  }

  testSorthTableByColumn(columnName: string, rowValues: Array<Array<string>>) {
    this.sortTableByColumn(columnName);
    const { rows } = this.getTableInfo();
    expect(rows).toHaveLength(rowValues.length);
    rowValues.forEach((row, index) => {
      const tds = rows![index].querySelectorAll('td');
      this.testRowValues(tds, row);
    });
  }

  testFilterTableByValue(
    filterIndex: number,
    value: string,
    rowValues: Array<Array<string>>,
  ) {
    const { rows } = this.applyFilterByValue(filterIndex, value);
    expect(rows).toHaveLength(rowValues.length);
    rowValues.forEach((row, index) => {
      const tds = rows![index].querySelectorAll('td');
      this.testRowValues(tds, row);
    });
  }
}
