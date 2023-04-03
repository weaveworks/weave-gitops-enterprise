import { useGetSecretDetails } from '../../../contexts/Secrets';
import { Routes } from '../../../utils/nav';
import { generateRowHeaders, SectionRowHeader } from '../../RowHeader';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import moment from 'moment';
import SecretDetailsTabs from './SecretDetailsTabs';
import { request } from '../../../utils/request';
import { SyncExternalSecretsResponse } from '../../../cluster-services/cluster_services.pb';
import { Button, theme } from '@weaveworks/weave-gitops';
import SyncIcon from '@material-ui/icons/Sync';

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

  const syncSecret = (payload: any): Promise<SyncExternalSecretsResponse> => {
    return request('POST', `/v1/external-secrets/sync`, {
      body: JSON.stringify(payload),
    });
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
        <Button
          id="create-secrets"
          startIcon={<SyncIcon />}
          style={{ marginBottom: theme.spacing.medium }}
          onClick={() =>
            syncSecret({
              clusterName,
              namespace,
              externalSecretName,
            })
          }
        >
          SYNC
        </Button>
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
