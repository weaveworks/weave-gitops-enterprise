import {
  PolicyViolationsList,
  RouterTab,
  SubRouterTabs,
  useFeatureFlags,
} from '@weaveworks/weave-gitops';
import { useRouteMatch } from 'react-router-dom';
import { Page } from '../Layout/App';
import PolicyAuditList from './Audit/PolicyAuditList';
import WarningMsg from './Audit/WarningMsg';
import { PoliciesTab } from './PoliciesListTab';

const Policies = () => {
  const { path } = useRouteMatch();
  const { isFlagEnabled } = useFeatureFlags();

  const isQueryServiceExplorerEnabled = isFlagEnabled(
    'WEAVE_GITOPS_FEATURE_EXPLORER',
  );
  const isQueryServiceBackendEnabled = isFlagEnabled(
    'WEAVE_GITOPS_FEATURE_QUERY_SERVICE_BACKEND',
  );

  return (
    <Page path={[{ label: 'Policies' }]}>
      <SubRouterTabs rootPath={`${path}/list`}>
        <RouterTab name="Policies" path={`${path}/list`}>
          <PoliciesTab />
        </RouterTab>
        <RouterTab name="Policy Audit" path={`${path}/audit`}>
          {isQueryServiceBackendEnabled && isQueryServiceExplorerEnabled ? (
            <PolicyAuditList />
          ) : (
            <WarningMsg />
          )}
        </RouterTab>
        <RouterTab name="Enforcement Events" path={`${path}/enforcement`}>
          <PolicyViolationsList req={{}} />
        </RouterTab>
      </SubRouterTabs>
    </Page>
  );
};

export default Policies;
