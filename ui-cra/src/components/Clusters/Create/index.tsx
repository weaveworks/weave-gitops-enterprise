import Grid from '@material-ui/core/Grid';
import { useParams } from 'react-router-dom';
import useClusters from '../../../hooks/clusters';
import useTemplates from '../../../hooks/templates';
import { ContentWrapper, Title } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import ClusterForm from '../Form';

const CreateClusterPage = () => {
  const clustersCount = useClusters().count;
  const { templateName } = useParams<{ templateName: string }>();
  const { getTemplate, isLoading } = useTemplates();
  return (
    <PageTemplate documentTitle="WeGo · Create new cluster">
      <SectionHeader
        className="count-header"
        path={[
          { label: 'Clusters', url: '/', count: clustersCount },
          { label: 'Create new cluster' },
        ]}
      />
      <ContentWrapper loading={isLoading}>
        <Grid container>
          <Grid item xs={12} sm={10} md={10} lg={8}>
            <Title>Create new cluster with template</Title>
          </Grid>
          <ClusterForm template={getTemplate(templateName)} />
        </Grid>
      </ContentWrapper>
    </PageTemplate>
  );
};

export default CreateClusterPage;
