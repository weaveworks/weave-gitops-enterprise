import { ListError } from '@weaveworks/progressive-delivery/api/prog/types.pb';
import {
  Flex,
  RouterTab,
  SubRouterTabs,
  YamlView,
  createYamlCommand,
} from '@weaveworks/weave-gitops';
import { Pipeline } from '../../../api/pipelines/types.pb';
import { useGetPipeline } from '../../../contexts/Pipelines';
import { Routes } from '../../../utils/nav';
import { Page } from '../../Layout/App';
import { NotificationsWrapper } from '../../Layout/NotificationsWrapper';
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
        <Flex column gap="16" wide>
          <SubRouterTabs rootPath={`${path}/status`}>
            <RouterTab name="Status" path={`${path}/status`}>
              <Workloads pipeline={data?.pipeline || ({} as Pipeline)} />
            </RouterTab>
            <RouterTab name="Yaml" path={`${path}/yaml`}>
              <YamlView
                yaml={data?.pipeline?.yaml || ''}
                header={createYamlCommand(
                  'Pipeline',
                  data?.pipeline?.name,
                  data?.pipeline?.namespace,
                )}
              />
            </RouterTab>
          </SubRouterTabs>
        </Flex>
      </NotificationsWrapper>
    </Page>
  );
};

export default PipelineDetails;
