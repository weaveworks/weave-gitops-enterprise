import React from 'react';
import { mount } from '@cypress/react';
import { withContext } from '../../src/utils/test-utils';
import ResponsiveDrawer from '../../src/components/ResponsiveDrawer';

import { ThemeProvider } from 'styled-components';
import { MuiThemeProvider } from '@material-ui/core';
import RequestContextProvider from '../../src/contexts/Request';
import { muiTheme } from '../../src/muiTheme';
import {
  AppContextProvider,
  applicationsClient,
  theme,
} from '@weaveworks/weave-gitops';
import { QueryClient, QueryClientProvider } from 'react-query';
import { MemoryRouter } from 'react-router-dom';
import { ProgressiveDeliveryProvider } from '../../src/contexts/ProgressiveDelivery';
import { ProgressiveDeliveryService } from '@weaveworks/progressive-delivery';
import { PipelinesProvider } from '../../src/contexts/Pipelines';
import { Pipelines } from '../../src/api/pipelines/pipelines.pb';
import { GlobalStyle } from '../../src/App';

const responses = {
  '/oauth2/userinfo': {
    email: 'wego-admin',
    groups: null,
  },
  'v1/applications/parse_repo_url*': {
    name: 'capd-demo-reloaded',
    provider: 'GitHub',
    owner: 'wkp-example-org',
  },
  'v1/config*': {
    repositoryURL: 'https://github.com/wkp-example-org/capd-demo-reloaded',
  },
  'v1/templates*': {
    templates: [
      {
        name: 'cluster-template-development',
        description: 'A simple CAPD template',
        provider: 'docker',
        parameters: [
          {
            name: 'CLUSTER_NAME',
            description: 'This is used for the cluster naming.',
            required: true,
            options: [],
            default: '',
          },
        ],
        objects: [
          {
            kind: 'GitopsCluster',
            apiVersion: 'gitops.weave.works/v1alpha1',
            parameters: ['CLUSTER_NAME', 'NAMESPACE'],
            name: '${CLUSTER_NAME}',
            displayName: '',
          },
        ],
        error: '',
        annotations: {},
        templateKind: 'CAPITemplate',
      },
    ],
    total: 5,
  },
  '/v1/clusters*': {
    gitopsClusters: [
      {
        name: 'management',
        namespace: '',
        annotations: {},
        labels: {},
        conditions: [
          {
            type: 'Ready',
            status: 'True',
            reason: '',
            message: '',
            timestamp: '',
          },
        ],
        capiClusterRef: null,
        secretRef: null,
        capiCluster: null,
        controlPlane: true,
      },
    ],
    total: 12,
    nextPageToken: '',
  },
  '/v1/enterprise/version': {
    version: '0.9.4-93-g3e77df4',
  },
  '/v1/featureflags*': {
    flags: {
      CLUSTER_USER_AUTH: 'true',
      OIDC_AUTH: 'true',
      WEAVE_GITOPS_FEATURE_CLUSTER: 'true',
      WEAVE_GITOPS_FEATURE_PIPELINES: 'true',
      WEAVE_GITOPS_FEATURE_TENANCY: 'true',
    },
  },
};

export const defaultContexts = () => [
  [ThemeProvider, { theme: theme }],
  [MuiThemeProvider, { theme: muiTheme }],
  [RequestContextProvider],
  [QueryClientProvider, { client: new QueryClient() }],
  [MemoryRouter, { initialEntries: ['/templates'] }],
  [ProgressiveDeliveryProvider, { api: ProgressiveDeliveryService }],
  [PipelinesProvider, { api: Pipelines }],
  [AppContextProvider, { applicationsClient }],
];

const page = [
  {
    name: 'templateHeader',
    select: `div[role="heading"] a[href="/templates"]`,
  },
  {
    name: 'templateCount',
    selectByXPath: `//*[@href="/templates"]/parent::div[@role="heading"]/following-sibling::div`,
  },
  { name: 'templateTiles', selectAll: `[data-template-name]` },
  { name: 'templatesList', selectAll: `#templates-list tbody tr` },
  { name: 'templateProvider', select: `#filter-by-provider` },
  {
    name: 'templateProviderPopup',
    selectAll: `ul#filter-by-provider-popup li`,
  },
  { name: 'templateView', selectAll: `#display-action > svg` },
];

const templatesSelectors = {
  page,
};

// const templateHeader = { name: "templateHeader", select: `div[role="heading"] a[href="/templates"]` };
// const templateCount = {
//   select: `//*[@href="/templates"]/parent::div[@role="heading"]/following-sibling::div`,
// };
// const templateTiles = { selectAll: `[data-template-name]` };
// const templatesList = { selectAll: `#templates-list tbody tr` };
// const templateProvider = { select: `#filter-by-provider` };
// const templateProviderPopup = { selectAll: `ul#filter-by-provider-popup li` };
// const templateView = { selectAll: `#display-action > svg` };

it('renders learn react link', () => {
  const wrap = withContext([...defaultContexts()]);
  for (const [url, response] of Object.entries(responses)) {
    cy.intercept('GET', url, response).as(`get ${url}`);
  }
  mount(
    wrap(
      <>
        <GlobalStyle />
        <ResponsiveDrawer />
      </>,
    ),
  );

  for (const selector of templatesSelectors.page) {
    if (selector.select) {
      cy.get(selector.select).should('exist');
    } else if (selector.selectAll) {
      cy.get(selector.selectAll).should('exist');
    } else if (selector.selectByXPath) {
      cy.xpath(selector.selectByXPath).should('exist');
    }
  }
});
