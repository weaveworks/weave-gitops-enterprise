import { FC } from 'react';
import { EnabledComponent } from '../../api/query/query.pb';
import { useIsEnabledForComponent } from '../../hooks/query';
import Explorer from '../Explorer/Explorer';
import WarningMsg from '../Explorer/WarningMsg';
import { Page } from '../Layout/App';

const ClusterDiscovery: FC = () => {
  const isExplorerEnabled = useIsEnabledForComponent(
    EnabledComponent.clusterdiscovery,
  );

  return (
    <Page
      path={[
        {
          label: 'Cluster Discovery',
        },
      ]}
    >
      {isExplorerEnabled ? (
        <Explorer category="clusterdiscovery" enableBatchSync={false} />
      ) : (
        <WarningMsg />
      )}
    </Page>
  );
};

export default ClusterDiscovery;
