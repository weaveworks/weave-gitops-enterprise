import React from 'react';
import { mount } from '@cypress/react';
import YAML from 'yamljs';

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
        parameters: [],
        objects: [],
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

interface Selector {
  select?: string;
  selectAll?: string;
  selectByXPath?: string;
}

type Selectors = {
  [group: string]: { [section: string]: { [name: string]: Selector } };
};

const select = (selector: Selector) => {
  if (selector.select) {
    return cy.get(selector.select).should('have.length', 1);
  } else if (selector.selectAll) {
    return cy.get(selector.selectAll).should('exist');
  } else if (selector.selectByXPath) {
    return cy.xpath(selector.selectByXPath).should('exist');
  }
};

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

  cy.readFile('../test/selectors/selectors.yaml')
    .then(yaml => YAML.parse(yaml) as Selectors)
    .then(({ templates: selectors }) => {
      for (const [selectorName, selector] of Object.entries(selectors.page)) {
        select(selector);
      }

      select(selectors.page.gridViewButton).click();
      select(selectors.gridView.tiles);
      select(selectors.gridView.provider).click();
      select(selectors.gridView.providerPopup);
    });
});
