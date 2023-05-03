import SyncIcon from '@material-ui/icons/Sync';
import { Button, theme } from '@weaveworks/weave-gitops';
import moment from 'moment';
import { useState } from 'react';
import useNotifications from '../../../contexts/Notifications';
import { useGetSecretDetails } from '../../../contexts/Secrets';
import { Routes } from '../../../utils/nav';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import { generateRowHeaders, SectionRowHeader } from '../../RowHeader';
import SecretDetailsTabs from './SecretDetailsTabs';
import { syncSecret } from './SyncSecret';
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
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const { setNotifications } = useNotifications();

  return (
    <PageTemplate
      documentTitle="Secrets"
      path={[
        { label: 'Secrets', url: Routes.Secrets },
        { label: secretDetails?.externalSecretName || '' },
      ]}
    >
      <ContentWrapper loading={isSecretDetailsLoading}>
        <Button
          id="create-secrets"
          loading={isLoading}
          startIcon={<SyncIcon />}
          style={{ marginBottom: theme.spacing.medium }}
          onClick={() => {
            syncSecret(
              {
                clusterName,
                namespace,
                externalSecretName,
              },
              setNotifications,
              setIsLoading,
            );
          }}
        >
          Sync
        </Button>
        <Box paddingBottom={3}>{generateRowHeaders(defaultHeaders)}</Box>
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
