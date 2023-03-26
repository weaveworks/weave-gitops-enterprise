import { PageTemplate } from '../Layout/PageTemplate';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { useListSecrets } from '../../contexts/Secrets';
import { SecretsTable } from './Table';
import { Button, Icon, IconType } from '@weaveworks/weave-gitops';
import { useHistory } from 'react-router-dom';
import { useCallback } from 'react';
import { Routes } from '../../utils/nav';

const SecretsList = () => {
  const { data, isLoading } = useListSecrets({});
  const history = useHistory();

  const handleCreateSecret = useCallback(
    (url: string) => history.push(url),
    [history],
  );

  return (
    <PageTemplate documentTitle="Secrets" path={[{ label: 'Secrets' }]}>
      <ContentWrapper loading={isLoading} errors={data?.errors}>
        <Button
          id="create-secrets"
          startIcon={<Icon type={IconType.AddIcon} size="base" />}
          onClick={() => handleCreateSecret(Routes.CreateSecret)}
        >
          CREATE EXTERNAL SECRET
        </Button>
        <Button
          id="create-sops-secrets"
          startIcon={<Icon type={IconType.AddIcon} size="base" />}
          onClick={() => handleCreateSecret(Routes.CreateSopsSecret)}
        >
          CREATE SOPS SECRET
        </Button>
        {data?.secrets && (
          <>
            <h2
              style={{
                margin: '32px 0 8px 0',
              }}
            >
              ExternalSecrets List
            </h2>
            <SecretsTable secrets={data.secrets} />
          </>
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default SecretsList;
