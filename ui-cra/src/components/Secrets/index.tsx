import { PageTemplate } from '../Layout/PageTemplate';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { useListSecrets } from '../../contexts/Secrets';
import { SecretsTable } from './Table';
import { Button, Icon, IconType } from '@weaveworks/weave-gitops';
import { useHistory } from 'react-router-dom';
import { useCallback } from 'react';

const SecretsList = () => {
  const { data, isLoading } = useListSecrets({});
  const history = useHistory();

  const handleCreateSecret = useCallback(
    () => history.push(`/secrets/create`),
    [history],
  );

  return (
    <PageTemplate documentTitle="Secrets" path={[{ label: 'Secrets' }]}>
      <ContentWrapper loading={isLoading} errors={data?.errors}>
        <Button
          id="create-cluster"
          startIcon={<Icon type={IconType.AddIcon} size="base" />}
          onClick={handleCreateSecret}
        >
          CREATE A SECRET
        </Button>
        {data?.secrets && <SecretsTable secrets={data.secrets} />}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default SecretsList;
