import { Button } from '@weaveworks/weave-gitops';
import moment from 'moment';
import { useState } from 'react';
import useNotifications from '../../../contexts/Notifications';
import { useGetSecretDetails } from '../../../contexts/Secrets';
import { Routes } from '../../../utils/nav';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import { generateRowHeaders, SectionRowHeader } from '../../RowHeader';
import SecretDetailsTabs from './SecretDetailsTabs';
import { useSyncSecret } from './SyncSecret';
import { Box } from '@material-ui/core';

const SecretDetails = ({
  externalSecretName,
  clusterName,
  namespace,
}: {
  externalSecretName: string;
  clusterName: string;
  namespace: string;
}) => {
  const { data: secretDetails, isLoading: isSecretDetailsLoading } =
    useGetSecretDetails({
      externalSecretName,
      clusterName,
      namespace,
    });
  const defaultHeaders: Array<SectionRowHeader> = [
    {
      rowkey: 'Status',
      value:
        secretDetails?.status === 'NotReady'
          ? 'Not Ready'
          : secretDetails?.status,
    },
    {
      rowkey: 'Last Updated',
      value: moment(secretDetails?.timestamp).fromNow(),
    },
  ];
  const [syncing, setSyncing] = useState(false);
  const { setNotifications } = useNotifications();

  const sync = useSyncSecret({
    clusterName,
    namespace,
    externalSecretName,
  });

  const handleSyncClick = () => {
    setSyncing(true);
    setNotifications([]);
    return sync()
      .catch(err => {
        setNotifications([
          {
            message: { text: err?.message },
            severity: 'error',
          },
        ]);
      })
      .finally(() => setSyncing(false));
  };

  return (
    <PageTemplate
      documentTitle="Secrets"
      path={[
        { label: 'Secrets', url: Routes.Secrets },
        { label: secretDetails?.externalSecretName || '' },
      ]}
    >
      <ContentWrapper loading={isSecretDetailsLoading}>
        <Box paddingBottom={3}>
          <Button
            id="sync-secret"
            loading={syncing}
            variant="outlined"
            onClick={handleSyncClick}
            style={{ marginRight: 0, textTransform: 'uppercase' }}
          >
            Sync
          </Button>
        </Box>
        {generateRowHeaders(defaultHeaders)}
        <SecretDetailsTabs
          externalSecretName={externalSecretName}
          clusterName={clusterName}
          namespace={namespace}
          secretDetails={secretDetails || {}}
        />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default SecretDetails;
