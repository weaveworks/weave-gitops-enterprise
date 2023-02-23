import { PageTemplate } from '../Layout/PageTemplate';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PolicyConfigsTable } from './Table';
import { useListPolicyConfigs } from '../../contexts/PolicyConfigs';

const PolicyConfigsList = () => {
  const { data, isLoading } = useListPolicyConfigs({});
  return (
    <PageTemplate documentTitle="PolicyConfigs" path={[{ label: 'PolicyConfigs' }]}>
      <ContentWrapper loading={isLoading} errors={data?.errors}>
        {data?.policyConfigs && <PolicyConfigsTable PolicyConfigs={data.policyConfigs} />}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default PolicyConfigsList;
