import { Flex, Link, Text } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { Pipeline, Strategy } from '../../../api/pipelines/types.pb';
import PromotePipeline from './PromotePipeline';
import PromotionInfo from './PromotionInfo';
import { EnvironmentCard } from './styles';
import Target from './Target';

const getEnvStrategy = (strategy?: Strategy) => {
  if (strategy?.pullRequest) {
    return {
      strategy: 'Pull Request',
      ...strategy?.pullRequest,
    };
  }
  if (strategy?.notification) {
    return {
      strategy: 'Notification',
    };
  }
  return {
    strategy: '-',
  };
};

const EnvironmentContainer = styled(Flex)`
  background: ${props => props.theme.colors.pipelineGray}};
  border-radius: 8px;
  padding: ${props => props.theme.spacing.small};
`;

const PromotionContainer = styled(Flex)`
  height: 54px;
  padding: ${props => props.theme.spacing.xs} 0;
`;

function Workloads({
  pipeline,
  className,
}: {
  pipeline: Pipeline;
  className?: string;
}) {
  const environments = pipeline?.environments || [];
  const targetsStatuses = pipeline?.status?.environments || {};

  return (
    <Flex gap="4" tall wide className={className}>
      {environments.map((env, index) => {
        const status = targetsStatuses[env.name!].targetsStatuses || [];
        const promoteVersion =
          targetsStatuses[env.name!].waitingStatus?.revision || '';
        const envStrategy = getEnvStrategy(env.promotion?.strategy);

        return (
          <EnvironmentContainer key={index} tall column gap="16">
            <PromotionInfo targets={status} />
            <EnvironmentCard background={index} column>
              <Flex column gap="8" wide>
                <Flex between align wide>
                  <Text bold capitalize size="large">
                    {env.name}
                  </Text>
                  <Text>{env.targets?.length || '0'} TARGETS</Text>
                </Flex>
                <Flex gap="8" wide start>
                  <Text bold>Strategy:</Text>
                  <Text> {envStrategy.strategy}</Text>
                </Flex>
              </Flex>
              <PromotionContainer column gap="12" wide>
                {envStrategy.strategy === 'Pull Request' && (
                  <Flex column gap="8" wide start>
                    <Flex gap="8" wide start>
                      <Text bold>Branch:</Text>
                      <Text> {envStrategy.branch}</Text>
                    </Flex>
                    <Flex gap="8" wide start>
                      <Text bold>URL:</Text>
                      <Link to={envStrategy.url}>{envStrategy.url}</Link>
                    </Flex>
                  </Flex>
                )}

                {env.promotion?.manual && index < environments.length - 1 && (
                  <PromotePipeline
                    req={{
                      name: pipeline.name,
                      env: env.name,
                      namespace: pipeline.namespace,
                      revision: promoteVersion,
                    }}
                    promoteVersion={promoteVersion || ''}
                  />
                )}
              </PromotionContainer>
            </EnvironmentCard>
            {status.map((target, indx) => (
              <Target key={indx} target={target} background={index} />
            ))}
          </EnvironmentContainer>
        );
      })}
    </Flex>
  );
}

export default styled(Workloads)`
  background: ${props => props.theme.colors.backGray}};
  padding: ${props => props.theme.spacing.medium};
  overflow-x: auto;
  box-sizing: border-box;
  //for PR url overflow
  ${Link} {
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
  }
`;
