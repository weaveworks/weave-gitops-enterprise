import { act, render, screen } from '@testing-library/react';
import moment from 'moment';
import { EnterpriseClientContext } from '../../../contexts/API';
import {
  SecretsClientMock,
  defaultContexts,
  withContext,
} from '../../../utils/test-utils';
import SecretDetails from '../SecretDetails';

const MockSecretResponse = {
  secretName: 'secret Name',
  externalSecretName: 'external Secret Name',
  clusterName: 'management',
  namespace: 'flux-system',
  secretStore: 'secret Store name',
  secretStoreType: 'secret Store Type',
  secretPath: 'secret Path',
  property: 'property',
  version: 'version',
  status: 'Ok',
  timestamp: '2022-07-30T11:23:55Z',
};

describe('SecretDetails', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: SecretsClientMock;

  beforeEach(() => {
    api = new SecretsClientMock();
    wrap = withContext([
      ...defaultContexts(),
      [EnterpriseClientContext.Provider, { value: { enterprise: api } }],
    ]);
  });

  it('renders get Secret details with tabs', async () => {
    const secret = MockSecretResponse;
    api.GetExternalSecretReturns = MockSecretResponse;
    await act(async () => {
      const c = wrap(
        <SecretDetails
          externalSecretName={secret.externalSecretName}
          clusterName={secret.clusterName}
          namespace={secret.namespace}
        />,
      );
      render(c);
    });
    //check Tabs
    expect(screen.getByTitle(secret.externalSecretName)).toBeTruthy();
    const tabs = screen.getAllByRole('tab');
    expect(secret.clusterName).toBeDefined();
    expect(tabs).toHaveLength(3);
    expect(tabs[0]).toHaveTextContent('Details');
    expect(tabs[1]).toHaveTextContent('Events');
    expect(tabs[2]).toHaveTextContent('Yaml');

    expect(screen.getByTestId('Status')).toHaveTextContent(secret.status);
    expect(screen.getByTestId('Last Updated')).toHaveTextContent(
      moment(secret.timestamp).fromNow(),
    );
  });

  it('renders secret details tab', async () => {
    const secret = MockSecretResponse;
    api.GetExternalSecretReturns = MockSecretResponse;
    await act(async () => {
      const c = wrap(
        <SecretDetails
          externalSecretName={secret.externalSecretName}
          clusterName={secret.clusterName}
          namespace={secret.namespace}
        />,
      );
      render(c);
    });
    expect(screen.getByTestId('External Secret')).toHaveTextContent(
      secret.externalSecretName,
    );
    expect(screen.getByTestId('K8s Secret')).toHaveTextContent(
      secret.secretName,
    );
    expect(screen.getByTestId('Cluster')).toHaveTextContent(secret.clusterName);
    expect(screen.getByTestId('Secret Store')).toHaveTextContent(
      secret.secretStore,
    );
    expect(screen.getByTestId('Secret Store Type')).toHaveTextContent(
      secret.secretStoreType,
    );
    expect(screen.getByTestId('Secret path')).toHaveTextContent(
      secret.secretPath,
    );
    expect(screen.getByTestId('Version')).toHaveTextContent(secret.version);
  });
});
