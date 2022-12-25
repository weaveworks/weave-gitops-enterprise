import { RouterTab } from '@weaveworks/weave-gitops';
import { useGetSecretEvent } from '../../../contexts/Secrets';
import { Routes } from '../../../utils/nav';
import { CustomSubRouterTabs } from '../../Workspaces/WorkspaceStyles';

const SecretDetailsTabs = ({
  clusterName,
  namespace,
  externalSecretName,
}: {
  clusterName: string;
  namespace: string;
  externalSecretName: string;
}) => {
  const path = Routes.SecretDetails;
  const { data: secretEvents, isLoading: isSecretEventsLoading } =
    useGetSecretEvent({
      clusterName,
      
    });
    console.log(secretEvents)
  return (
    <div style={{ minHeight: 'calc(100vh - 335px)' }}>
      <CustomSubRouterTabs rootPath={`${path}/secretDetails`}>
        <RouterTab name="Details" path={`${path}/secretDetails`}>
          <> </>
        </RouterTab>

        <RouterTab name="Events" path={`${path}/events`}>
          <></>
        </RouterTab>
      </CustomSubRouterTabs>
    </div>
  );
};

export default SecretDetailsTabs;
