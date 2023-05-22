import { PageTemplate } from '../Layout/PageTemplate';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PolicyConfigsTable } from './Table';
import { useListPolicyConfigs } from '../../contexts/PolicyConfigs';
import { Button, Icon, IconType } from '@weaveworks/weave-gitops';
import { useNavigate } from 'react-router-dom';
import { useCallback } from 'react';

const PolicyConfigsList = () => {
  const { data, isLoading } = useListPolicyConfigs({});
  const navigate = useNavigate();

  const handleCreateSecret = useCallback(
    () => navigate(`/policyConfigs/create`),
    [navigate],
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
          CREATE A POLICY CONFIG
        </Button>
        {data?.policyConfigs && (
          <PolicyConfigsTable PolicyConfigs={data.policyConfigs} />
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default PolicyConfigsList;
