import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PolicyViolationsList } from './Table';
import { useState } from 'react';
import { ListPolicyValidationsResponse } from '../../cluster-services/cluster_services.pb';
import useClusters from './../../contexts/Clusters';

const PoliciesViolations = () => {
  const clustersCount = useClusters().count;
  const [data, setData] = useState<ListPolicyValidationsResponse>();

  const onSuccess = (dt: ListPolicyValidationsResponse) => {
    setData(dt);
  };

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Violation Log">
        <SectionHeader
          className="count-header"
          path={[
            { label: 'Clusters', url: '/clusters', count: clustersCount },
            {
              label: 'Violation Log',
              url: 'violations',
              count: data?.total,
            },
          ]}
        />
        <ContentWrapper errors={data?.errors}>
          <PolicyViolationsList onSuccess={onSuccess} req={{}} />
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default PoliciesViolations;
