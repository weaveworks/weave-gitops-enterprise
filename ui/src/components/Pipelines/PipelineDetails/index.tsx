import { Box } from '@material-ui/core';
import { ListError } from '@weaveworks/progressive-delivery/api/prog/types.pb';
import {
  Flex,
  RouterTab,
  SubRouterTabs,
  YamlView,
} from '@weaveworks/weave-gitops';
import { Pipeline } from '../../../api/pipelines/types.pb';
import { useGetPipeline } from '../../../contexts/Pipelines';
import { Routes } from '../../../utils/nav';
import { Page } from '../../Layout/App';
import { NotificationsWrapper } from '../../Layout/NotificationsWrapper';
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
  const path = `/pipelines/details`;

  return (
    <Page
      loading={isLoading}
      path={[
        {
          label: 'Pipelines',
          url: Routes.Pipelines,
        },
        {
          label: name,
        },
      ]}
    >
      <NotificationsWrapper
        errors={mappedErrors(data?.errors || [], namespace)}
      >
        <Box marginBottom={2}>
          <Flex wide end>
            <EditButton
              className={classes.editButton}
              resource={data?.pipeline || ({} as Pipeline)}
            />
          </Flex>
        </Box>
        <SubRouterTabs rootPath={`${path}/status`}>
          <RouterTab name="Status" path={`${path}/status`}>
            <Workloads pipeline={data?.pipeline || ({} as Pipeline)} />
          </RouterTab>
          <RouterTab name="Yaml" path={`${path}/yaml`}>
            <YamlView
              yaml={data?.pipeline?.yaml || ''}
              object={{
                kind: 'Pipeline',
                name: data?.pipeline?.name,
                namespace: data?.pipeline?.namespace,
              }}
            />
          </RouterTab>
          <RouterTab name="Pull Requests" path={`${path}/pullrequests`}>
            <PipelinePullRequests pipeline={data?.pipeline} />
          </RouterTab>
        </SubRouterTabs>
      </NotificationsWrapper>
    </Page>
  );
};

export default PipelineDetails;
