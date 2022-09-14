import Grid from '@material-ui/core/Grid';
import { FC } from 'react';
import { Redirect, useParams } from 'react-router-dom';
import useClusters from '../../../hooks/clusters';
import useTemplates from '../../../hooks/templates';
import { GitopsClusterEnriched } from '../../../types/custom';
import { ContentWrapper, Title } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import ClusterForm from '../Form';
import { getCreateRequestAnnotation } from '../Form/utils';

const EditCluster: FC<{ cluster?: GitopsClusterEnriched | null }> = ({
  cluster,
}) => {
  const { getTemplate } = useTemplates();

  const templateName =
    cluster && getCreateRequestAnnotation(cluster)?.template_name;
  if (!templateName) {
    return <Redirect to="/clusters" />;
  }

  return <ClusterForm template={getTemplate(templateName)} cluster={cluster} />;
};

const EditClusterPage = () => {
  const { count: clustersCount, isLoading, getCluster } = useClusters();
  const { clusterName } = useParams<{ clusterName: string }>();
  const { isLoading: isTemplateLoading } = useTemplates();
  return (
    <PageTemplate documentTitle="WeGo · Create new cluster">
      <SectionHeader
        className="count-header"
        path={[
          { label: 'Clusters', url: '/', count: clustersCount },
          { label: clusterName },
        ]}
      />
      <ContentWrapper loading={isLoading || isTemplateLoading}>
        <Grid container>
          <Grid item xs={12} sm={10} md={10} lg={8}>
            <Title>Edit cluster</Title>
          </Grid>
          <EditCluster cluster={getCluster(clusterName)} />
        </Grid>
      </ContentWrapper>
    </PageTemplate>
  );
};

export default EditClusterPage;
