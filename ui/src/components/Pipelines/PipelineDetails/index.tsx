import { ListError } from '@weaveworks/progressive-delivery/api/prog/types.pb';
import {
  Flex,
  RouterTab,
  SubRouterTabs,
  YamlView,
  createYamlCommand,
} from '@weaveworks/weave-gitops';
import { GetPipelineResponse } from '../../../api/pipelines/pipelines.pb';
import { Pipeline } from '../../../api/pipelines/types.pb';
import { useGetPipeline } from '../../../contexts/Pipelines';
import { Routes } from '../../../utils/nav';
import KeyValueTable, { KeyValuePairs } from '../../KeyValueTable';
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
const pipelineStrategyText = (data?: GetPipelineResponse): KeyValuePairs => {
  const pairs: KeyValuePairs = [];

  if (!data) {
    return pairs;
  }

  // Trying to differentiate between null and negative values
  if (data?.pipeline?.promotion?.manual === null) {
    pairs.push(['Automated', null]);
  } else {
    pairs.push([
      'Automated',
      data?.pipeline?.promotion?.manual ? 'False' : 'True',
    ]);
  }

  const strat = data?.pipeline?.promotion?.strategy;

  if (strat?.pullRequest === null && strat?.notification === null) {
    pairs.push(['Strategy', null]);
  } else {
    pairs.push([
      'Strategy',
      data?.pipeline?.promotion?.strategy?.pullRequest
        ? 'Pull Request'
        : 'Notification',
    ]);

    if (strat?.pullRequest !== null) {
      const pr = data.pipeline?.promotion?.strategy?.pullRequest;
      pairs.push(['URL', pr?.url]);
      pairs.push(['Branch', pr?.branch]);
    }
  }

  return pairs;
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
          <KeyValueTable pairs={pipelineStrategyText(data)} />

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
