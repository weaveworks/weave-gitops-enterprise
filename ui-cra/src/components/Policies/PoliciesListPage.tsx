import {
  PolicyViolationsList,
  RouterTab,
  SubRouterTabs,
} from '@weaveworks/weave-gitops';
import { useRouteMatch } from 'react-router-dom';
import { Page } from '../Layout/App';
import { PoliciesTab } from './PoliciesListTab';
import PolicyAuditList from './PolicyAuditList';

const Policies = () => {
  const { path } = useRouteMatch();

  return (
    <Page path={[{ label: 'Policies' }]}>
      <SubRouterTabs rootPath={`${path}/policiesList`}>
        <RouterTab name="Policies" path={`${path}/policiesList`}>
          <PoliciesTab />
        </RouterTab>
        <RouterTab name="Policy Audit" path={`${path}/policyAudit`}>
          <PolicyAuditList />
        </RouterTab>
        <RouterTab name="Enforcement Events" path={`${path}/enforcementEvent`}>
          <PolicyViolationsList req={{}} />
        </RouterTab>
      </SubRouterTabs>
    </Page>
  );
};

export default Policies;
