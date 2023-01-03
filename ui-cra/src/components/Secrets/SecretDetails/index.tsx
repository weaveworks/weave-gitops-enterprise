import { useGetSecretDetails } from '../../../contexts/Secrets';
import { Routes } from '../../../utils/nav';
import { generateRowHeaders, SectionRowHeader } from '../../RowHeader';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import moment from 'moment';
import SecretDetailsTabs from './SecretDetailsTabs';

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

  return (
      <PageTemplate
        documentTitle="Secrets"
        path={[
          { label: 'Secrets', url: Routes.Secrets },
          { label: secretDetails?.externalSecretName || '' },
        ]}
      >
        <ContentWrapper loading={isSecretDetailsLoading}>
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
