import { RouterTab, SubRouterTabs } from '@weaveworks/weave-gitops';
import { Pipeline } from '../../../api/pipelines/types.pb';
import { useGetPipeline } from '../../../contexts/Pipelines';
import { Routes } from '../../../utils/nav';
import CodeView from '../../CodeView';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';

import { ListError } from '@weaveworks/progressive-delivery/api/prog/types.pb';
import { EditButton } from './../../../components/Templates/Edit/EditButton';
import PipelinePullRequests from './PipelinePullRequests';
import { usePipelineStyles } from './styles';
import Workloads from './Workloads';

const mappedErrors = (
  errors: Array<string>,
  namespace: string,
): Array<ListError> => {
  return errors.map(err => ({
    message: err,
    namespace,
  }));
};
interface Props {
  name: string;
  namespace: string;
}

const PipelineDetails = ({ name, namespace }: Props) => {
  const { isLoading, data } = useGetPipeline({
    name,
    namespace,
  });

  const classes = usePipelineStyles();
  const path = `/applications/pipelines/details`;

  return (
    <PageTemplate
      documentTitle="Pipeline Details"
      path={[
        {
          label: 'Applications',
          url: Routes.Applications,
        },
        {
          label: 'Pipelines',
          url: Routes.Pipelines,
        },
        {
          label: name,
        },
      ]}
    >
      <ContentWrapper
        loading={isLoading}
        errors={mappedErrors(data?.errors || [], namespace)}
      >
        <EditButton
          className={classes.editButton}
          resource={data?.pipeline || ({} as Pipeline)}
        />
        <SubRouterTabs rootPath={`${path}/status`}>
          <RouterTab name="Status" path={`${path}/status`}>
            <Workloads pipeline={data?.pipeline || ({} as Pipeline)} />
          </RouterTab>
          <RouterTab name="Yaml" path={`${path}/yaml`}>
            <CodeView
              kind="Pipeline"
              code={data?.pipeline?.yaml || ''}
              object={data?.pipeline || {}}
            />
          </RouterTab>
          <RouterTab name="Pull Requests" path={`${path}/pullrequests`}>
            <PipelinePullRequests pipeline={data?.pipeline} />
          </RouterTab>
        </SubRouterTabs>
      </ContentWrapper>
    </PageTemplate>
  );
};

export default PipelineDetails;
