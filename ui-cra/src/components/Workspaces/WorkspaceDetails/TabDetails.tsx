import { RouterTab } from '@weaveworks/weave-gitops';
import { Routes } from '../../../utils/nav';
import { CustomSubRouterTabs } from '../WorkspaceStyles';
import { PoliciesTab } from './Tabs/Policies';
import { RoleBindingsTab } from './Tabs/RoleBindings';
import { RolesTab } from './Tabs/Roles';
import { ServiceAccountsTab } from './Tabs/ServiceAccounts';

const TabDetails = ({
  clusterName,
  workspaceName,
}: {
  clusterName: string;
  workspaceName: string;
}) => {
  const path = Routes.WorkspaceDetails;

  return (
    <div style={{ minHeight: 'calc(100vh - 335px)' }}>
      <CustomSubRouterTabs>
        <RouterTab name="Service Accounts" path={`/serviceAccounts`}>
          <ServiceAccountsTab
            clusterName={clusterName}
            workspaceName={workspaceName}
          />
        </RouterTab>

        <RouterTab name="Roles" path={`/roles`}>
          <RolesTab clusterName={clusterName} workspaceName={workspaceName} />
        </RouterTab>

        <RouterTab name="Role Bindings" path={`/roleBindings`}>
          <RoleBindingsTab
            clusterName={clusterName}
            workspaceName={workspaceName}
          />
        </RouterTab>

        <RouterTab name="Policies" path={`/policies`}>
          <PoliciesTab
            clusterName={clusterName}
            workspaceName={workspaceName}
          />
        </RouterTab>
      </CustomSubRouterTabs>
    </div>
  );
};

export default TabDetails;
