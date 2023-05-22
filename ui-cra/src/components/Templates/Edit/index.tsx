import Grid from '@material-ui/core/Grid';
import { Kind, useGetObject } from '@weaveworks/weave-gitops';
import { FC } from 'react';
import { useGetPipeline } from '../../../contexts/Pipelines';
import { useGetTerraformObjectDetail } from '../../../contexts/Terraform';
import useClusters from '../../../hooks/clusters';
import useTemplates from '../../../hooks/templates';
import { Redirect, Routes } from '../../../utils/nav';
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
    return null;
    // <Redirect
    //   to={{
    //     pathname: Routes.Clusters,
    //     state: {
    //       notification: [
    //         {
    //           message: {
    //             text: 'No edit information is available for this resource.',
    //           },
    //           severity: 'error',
    //         },
    //       ],
    //     },
    //   }}
    // />
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
