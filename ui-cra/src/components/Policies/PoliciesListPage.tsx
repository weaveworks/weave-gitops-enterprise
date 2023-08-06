import {
  PolicyViolationsList,
  RouterTab,
  SubRouterTabs,
} from '@weaveworks/weave-gitops';
import { useRouteMatch } from 'react-router-dom';
import { Page } from '../Layout/App';
import { PoliciesTab } from './PoliciesListTab';
import PolicyAuditList from './Audit/PolicyAuditList';

const Policies = () => {
  const { path } = useRouteMatch();

  return (
    <Page path={[{ label: 'Policies' }]}>
      <SubRouterTabs rootPath={`${path}/list`}>
        <RouterTab name="Policies" path={`${path}/list`}>
          <PoliciesTab />
        </RouterTab>
        <RouterTab name="Policy Audit" path={`${path}/audit`}>
          <PolicyAuditList />
        </RouterTab>
        <RouterTab name="Enforcement Events" path={`${path}/enforcement`}>
          <PolicyViolationsList req={{}} />
        </RouterTab>
      </SubRouterTabs>
    </Page>
  );
};

export default Policies;
