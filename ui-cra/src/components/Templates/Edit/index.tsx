import Grid from '@material-ui/core/Grid';
import { Kind, Page, useGetObject } from '@weaveworks/weave-gitops';
import { FC } from 'react';
import { Redirect } from 'react-router-dom';
import { useGetPipeline } from '../../../contexts/Pipelines';
import { useGetTerraformObjectDetail } from '../../../contexts/Terraform';
import useClusters from '../../../hooks/clusters';
import useTemplates from '../../../hooks/templates';
import { Routes } from '../../../utils/nav';
import { Title } from '../../Layout/ContentWrapper';
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
        kind !== 'GitopsCluster' && kind !== 'Terraform' && kind !== 'Pipeline',
    },
  );

  const { getCluster } = useClusters();
  const cluster = getCluster(name);

  const { data: tfData } = useGetTerraformObjectDetail(
    {
      name,
      namespace,
      clusterName,
    },
    kind === 'Terraform',
  );

  const { data: pipelineData } = useGetPipeline(
    {
      name,
      namespace,
    },
    kind === 'Pipeline',
  );

  const getEditableResource = () => {
    switch (kind) {
      case 'Terraform':
        return tfData;
      case 'Pipeline':
        return pipelineData?.pipeline;
      case 'GitopsCluster':
        return cluster;
      default:
        return resource;
    }
  };

  return (
    <Page
      loading={isLoading || isTemplateLoading}
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
      <Grid container>
        <Grid item xs={12} sm={10} md={10} lg={8}>
          <Title>Edit resource</Title>
        </Grid>
        <EditResource resource={getEditableResource() || {}} />
      </Grid>
    </Page>
  );
};

export default EditResourcePage;
