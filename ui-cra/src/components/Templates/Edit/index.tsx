import Grid from '@material-ui/core/Grid';
import { Kind, useGetObject } from '@weaveworks/weave-gitops';
import { Automation, Source } from '@weaveworks/weave-gitops/ui/lib/objects';
import { FC } from 'react';
import { Redirect } from 'react-router-dom';
import { Pipeline } from '../../../api/pipelines/types.pb';
import {
  GetTerraformObjectResponse,
  Terraform,
} from '../../../api/terraform/terraform.pb';
import { useGetTerraformObjectDetail } from '../../../contexts/Terraform';
import useClusters from '../../../hooks/clusters';
import useTemplates from '../../../hooks/templates';
import { GitopsClusterEnriched } from '../../../types/custom';
import { Routes } from '../../../utils/nav';
import { ContentWrapper, Title } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import ResourceForm from '../Form';
import { getCreateRequestAnnotation } from '../Form/utils';

const EditResource: FC<{
  resource:
    | GitopsClusterEnriched
    | Automation
    | Source
    | GetTerraformObjectResponse
    | Pipeline;
}> = ({ resource }) => {
  const { getTemplate } = useTemplates();

  const templateName = getCreateRequestAnnotation(resource)?.template_name;

  console.log(getCreateRequestAnnotation(resource));

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

type Props = {
  name: string;
  namespace: string;
  kind: string;
  clusterName: string;
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

  const { data } = useGetTerraformObjectDetail({
    name,
    namespace,
    clusterName,
  });
  // const { object, yaml } = (data || {}) as GetTerraformObjectResponse;

  // console.log(object);

  // const editableResource =
  // kind === 'GitopsCluster'
  // ? (cluster as GitopsClusterEnriched)
  // : (resource as Automation | Source)

  // send data for Terraform

  return (
    <PageTemplate
      documentTitle="Edit resource"
      path={[
        { label: 'Resource' },
        {
          label:
            kind === 'GitopsCluster'
              ? cluster?.name || ''
              : resource?.name || data?.object?.name || '',
        },
      ]}
    >
      <ContentWrapper loading={isLoading || isTemplateLoading}>
        <Grid container>
          <Grid item xs={12} sm={10} md={10} lg={8}>
            <Title>Edit resource</Title>
          </Grid>
          <EditResource
            // resource={
            //   kind === 'GitopsCluster'
            //     ? (cluster as GitopsClusterEnriched)
            //     : (resource as Automation | Source)
            // }
            resource={data || {}}
          />
        </Grid>
      </ContentWrapper>
    </PageTemplate>
  );
};

export default EditResourcePage;
