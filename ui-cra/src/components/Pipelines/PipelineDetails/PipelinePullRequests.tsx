import { DataTable, Flex } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import styled from 'styled-components';
import { Pipeline } from '../../../api/pipelines/types.pb';
import { useGetPullRequestsForPipeline } from '../../../contexts/Pipelines';

type Props = {
  className?: string;
  pipeline?: Pipeline;
};

function PipelinePullRequests({ className, pipeline }: Props) {
  const { data: prs } = useGetPullRequestsForPipeline(pipeline);

  const rows = _.map(prs?.pullRequests, (url, env) => ({ env, url }));

  return (
    <div className={className}>
      <Flex wide>
        <DataTable
          fields={[
            { label: 'Environment', value: 'env' },
            {
              label: 'URL',
              value: r => (
                <a href={r.url} target="_blank">
                  {r.url}
                </a>
              ),
            },
          ]}
          rows={rows}
        />
      </Flex>
    </div>
  );
}

export default styled(PipelinePullRequests).attrs({
  className: PipelinePullRequests.name,
})`
  width: 100%;
`;
