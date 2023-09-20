import { useListSecrets } from '../../contexts/Secrets';
import { Routes } from '../../utils/nav';
import { Page } from '../Layout/App';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import { SecretsTable } from './Table';
import { Button, Flex, Icon, IconType, Text } from '@weaveworks/weave-gitops';
import { useCallback } from 'react';
import { useHistory } from 'react-router-dom';

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
        <Flex column gap="32">
          <Flex gap="12">
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
            <Flex column wide>
              <Text titleHeight semiBold size="large">
                ExternalSecrets List
              </Text>
              <SecretsTable secrets={data.secrets} />
            </Flex>
          )}
        </Flex>
      </NotificationsWrapper>
    </Page>
  );
};

export default SecretsList;
