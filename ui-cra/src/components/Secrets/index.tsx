import { PageTemplate } from '../Layout/PageTemplate';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { useListSecrets } from '../../contexts/Secrets';
import { SecretsTable } from './Table';

const SecretsList = () => {
  const { data, isLoading } = useListSecrets({});
  return (
    <PageTemplate
      documentTitle="Secrets"
      path={[{ label: 'Secrets'}]}
    >
      <ContentWrapper loading={isLoading} errors={data?.errors}>
        {data?.secrets && <SecretsTable secrets={data.secrets} />}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default SecretsList;
