import { Button, Icon, IconType } from '@weaveworks/weave-gitops';
import { useCallback } from 'react';
import { useHistory } from 'react-router-dom';
import { useListPolicyConfigs } from '../../contexts/PolicyConfigs';
import { Page } from '../Layout/App';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import { PolicyConfigsTable } from './Table';

const PolicyConfigsList = () => {
  const { data, isLoading } = useListPolicyConfigs({});
  const history = useHistory();

  const handleCreatePolicyConfig = useCallback(
    () => history.push(`/policyConfigs/create`),
    [history],
  );
  return (
    <Page loading={isLoading} path={[{ label: 'PolicyConfigs' }]}>
      <NotificationsWrapper errors={data?.errors}>
        <Button
          id="create-policy-config"
          startIcon={<Icon type={IconType.AddIcon} size="base" />}
          onClick={handleCreatePolicyConfig}
        >
          CREATE A POLICY CONFIG
        </Button>
        {data?.policyConfigs && (
          <PolicyConfigsTable PolicyConfigs={data.policyConfigs} />
        )}
      </NotificationsWrapper>
    </Page>
  );
};

export default PolicyConfigsList;
