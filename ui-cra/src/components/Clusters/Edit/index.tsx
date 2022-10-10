import Grid from '@material-ui/core/Grid';
import { FC } from 'react';
import { useParams, Redirect } from 'react-router-dom';
import useClusters from '../../../hooks/clusters';
import useTemplates from '../../../hooks/templates';
import { GitopsClusterEnriched } from '../../../types/custom';
import { ContentWrapper, Title } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import ResourceForm from '../Form';
import { getCreateRequestAnnotation } from '../Form/utils';

const EditResource: FC<{ resource?: any | null }> = ({ resource }) => {
  const { getTemplate } = useTemplates();

  const templateName =
    resource && getCreateRequestAnnotation(resource)?.template_name;

  if (!templateName) {
    return (
      <Redirect
        to={{
          pathname: '/clusters',
          state: {
            notification: [
              {
                message: {
                  text: 'No edit information is available for this resource.',
                },
                variant: 'danger',
              },
            ],
          },
        }}
      />
    );
  }

  return (
    <ResourceForm template={getTemplate(templateName)} resource={resource} />
  );
};

const EditClusterPage = () => {
  const { isLoading, getCluster } = useClusters();
  const { clusterName } = useParams<{ clusterName: string }>();
  const { isLoading: isTemplateLoading } = useTemplates();
  return (
    <PageTemplate
      documentTitle="Edit resource"
      path={[{ label: 'Clusters', url: '/' }, { label: clusterName }]}
    >
      <ContentWrapper loading={isLoading || isTemplateLoading}>
        <Grid container>
          <Grid item xs={12} sm={10} md={10} lg={8}>
            <Title>Edit resource</Title>
          </Grid>
          <EditResource resource={getCluster(clusterName)} />
        </Grid>
      </ContentWrapper>
    </PageTemplate>
  );
};

export default EditClusterPage;
