import { useListSecrets } from '../../contexts/Secrets';
import { SecretsTable } from './Table';
import { Button, Flex, Icon, IconType, Page } from '@weaveworks/weave-gitops';
import { useHistory } from 'react-router-dom';
import { useCallback } from 'react';
import { Routes } from '../../utils/nav';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';

const SecretsList = () => {
  const { data, isLoading } = useListSecrets({});
  const history = useHistory();

  const handleCreateSecret = useCallback(
    (url: string) => history.push(url),
    [history],
  );

  return (
    <Page loading={isLoading} path={[{ label: 'Secrets' }]}>
      <NotificationsWrapper errors={data?.errors}>
        <Flex center between>
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
        </Flex>
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
      </NotificationsWrapper>
    </Page>
  );
};

export default SecretsList;
