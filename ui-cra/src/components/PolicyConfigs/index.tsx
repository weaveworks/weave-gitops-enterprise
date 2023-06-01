import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import { PolicyConfigsTable } from './Table';
import { useListPolicyConfigs } from '../../contexts/PolicyConfigs';
import { Button, Icon, IconType, Page } from '@weaveworks/weave-gitops';
import { useHistory } from 'react-router-dom';
import { useCallback } from 'react';

const PolicyConfigsList = () => {
  const { data, isLoading } = useListPolicyConfigs({});
  const history = useHistory();

  const handleCreateSecret = useCallback(
    () => history.push(`/policyConfigs/create`),
    [history],
  );
  return (
    <Page loading={isLoading} path={[{ label: 'PolicyConfigs' }]}>
      <NotificationsWrapper errors={data?.errors}>
        <Button
          id="create-cluster"
          startIcon={<Icon type={IconType.AddIcon} size="base" />}
          onClick={handleCreateSecret}
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
