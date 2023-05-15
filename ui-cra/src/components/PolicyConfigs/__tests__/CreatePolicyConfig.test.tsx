import { act, render, screen } from '@testing-library/react';
import EnterpriseClientProvider from '../../../contexts/EnterpriseClient/Provider';
import {
  PolicyConfigsClientMock,
  defaultContexts,
  withContext,
} from '../../../utils/test-utils';
import CreatePolicyConfig from '../create';

describe('CreatePolicyConfig', () => {
  let wrap: (el: JSX.Element) => JSX.Element;
  let api: PolicyConfigsClientMock;

  beforeEach(() => {
    api = new PolicyConfigsClientMock();

    wrap = withContext([
      ...defaultContexts(),
      [EnterpriseClientProvider, { api }],
    ]);
  });
  it('renders create policyConfig form fields', async () => {
    await act(async () => {
      const c = wrap(<CreatePolicyConfig />);
      render(c);
    });

    expect(await screen.findByText('Create New PolicyConfig')).toBeTruthy();
    expect(
      document.querySelector("input[name='policyConfigName']"),
    ).toBeInTheDocument();

    expect(
      document.querySelector("input[name='clusterName']"),
    ).toBeInTheDocument();

    expect(
      document.querySelector("input[name='matchType']"),
    ).toBeInTheDocument();

    expect(
      document.querySelector("input[name='policies']"),
    ).toBeInTheDocument();
  });
});
