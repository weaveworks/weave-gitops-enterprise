import Grid from '@material-ui/core/Grid';
import { Automation, Source } from '@weaveworks/weave-gitops/ui/lib/objects';
import { FC } from 'react';
import { Redirect } from 'react-router-dom';
import useClusters from '../../../hooks/clusters';
import useTemplates from '../../../hooks/templates';
import { GitopsClusterEnriched } from '../../../types/custom';
import { Routes } from '../../../utils/nav';
import { ContentWrapper, Title } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import ResourceForm from '../Form';
import { getCreateRequestAnnotation } from '../Form/utils';

const EditResource: FC<{
  resource?: GitopsClusterEnriched | Automation | Source;
}> = ({ resource }) => {
  const { getTemplate } = useTemplates();

  const templateName =
    resource && getCreateRequestAnnotation(resource)?.template_name;

  if (!templateName) {
    return (
      <Redirect
        to={{
          pathname: Routes.Clusters,
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

const EditResourcePage: FC<{
  location: {
    state: { resource: GitopsClusterEnriched | Automation | Source };
  };
}> = ({ location }) => {
  const resource = location.state?.resource;
  const { isLoading } = useClusters();
  const { isLoading: isTemplateLoading } = useTemplates();

  return (
    <PageTemplate
      documentTitle="Edit resource"
      path={[
        { label: 'Resource', url: Routes.Clusters },
        { label: resource?.name },
      ]}
    >
      <ContentWrapper loading={isLoading || isTemplateLoading}>
        <Grid container>
          <Grid item xs={12} sm={10} md={10} lg={8}>
            <Title>Edit resource</Title>
          </Grid>
          <EditResource resource={resource} />
        </Grid>
      </ContentWrapper>
    </PageTemplate>
  );
};

export default EditResourcePage;
