import Grid from '@material-ui/core/Grid';
import { Kind, useGetObject } from '@weaveworks/weave-gitops';
import { FC } from 'react';
import { Redirect } from 'react-router-dom';
import { useGetPipeline } from '../../../contexts/Pipelines';
import { useGetTerraformObjectDetail } from '../../../contexts/Terraform';
import useClusters from '../../../hooks/clusters';
import useTemplates from '../../../hooks/templates';
import { Routes } from '../../../utils/nav';
import { ContentWrapper, Title } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import ResourceForm from '../Form';
import { getCreateRequestAnnotation } from '../Form/utils';
import { Resource } from './EditButton';

const EditResource: FC<{
  resource: Resource;
}> = ({ resource }) => {
  const { getTemplate } = useTemplates();

  const templateName = getCreateRequestAnnotation(resource)?.template_name;

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
                severity: 'error',
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

type Props = {
  name: string;
  namespace: string;
  clusterName: string;
  kind?: string;
};

const EditResourcePage: FC<Props> = props => {
  const { isLoading: isTemplateLoading } = useTemplates();
  const { name, namespace, kind, clusterName } = props;
  const { data: resource, isLoading } = useGetObject(
    name,
    namespace,
    kind as Kind,
    clusterName,
    {
      enabled:
        kind ===
        ('GitRepository' ||
          'Bucket' ||
          'HelmRepository' ||
          'HelmChart' ||
          'Kustomization' ||
          'HelmRelease' ||
          'OCIRepository'),
    },
  );

  const { getCluster } = useClusters();
  const cluster = getCluster(name);

  const { data: tfData } = useGetTerraformObjectDetail({
    name,
    namespace,
    clusterName,
  });

  const { data: pipelineData } = useGetPipeline({
    name,
    namespace,
  });

  const getEditableResource = () => {
    if (tfData) {
      return tfData;
    }
    if (pipelineData) {
      // remove type before merging, only using it to work with demo-01
      return { ...pipelineData.pipeline, type: 'Pipeline' };
    }
    if (kind === 'GitopsCluster') {
      return cluster;
    }
    return resource;
  };

  return (
    <PageTemplate
      documentTitle="Edit resource"
      path={[
        { label: 'Resource' },
        {
          label:
            cluster?.name ||
            resource?.name ||
            tfData?.object?.name ||
            pipelineData?.pipeline?.name ||
            '',
        },
      ]}
    >
      <ContentWrapper loading={isLoading || isTemplateLoading}>
        <Grid container>
          <Grid item xs={12} sm={10} md={10} lg={8}>
            <Title>Edit resource</Title>
          </Grid>
          <EditResource resource={getEditableResource() || {}} />
        </Grid>
      </ContentWrapper>
    </PageTemplate>
  );
};

export default EditResourcePage;
