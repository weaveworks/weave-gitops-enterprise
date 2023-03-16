import { PageTemplate } from '../Layout/PageTemplate';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PolicyConfigsTable } from './Table';
import { useListPolicyConfigs } from '../../contexts/PolicyConfigs';
import { Button, Icon, IconType } from '@weaveworks/weave-gitops';
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
    <PageTemplate
      documentTitle="PolicyConfigs"
      path={[{ label: 'PolicyConfigs' }]}
    >
      <ContentWrapper loading={isLoading} errors={data?.errors}>
        <Button
          id="create-cluster"
          startIcon={<Icon type={IconType.AddIcon} size="base" />}
          onClick={handleCreateSecret}
        >
          CREATE A PolicyConfig
        </Button>
        {data?.policyConfigs && (
          <PolicyConfigsTable PolicyConfigs={data.policyConfigs} />
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default PolicyConfigsList;
