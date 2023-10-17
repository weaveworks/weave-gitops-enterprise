import { Flex, Text } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import styled from 'styled-components';
import { Pipeline, Promotion } from '../../../api/pipelines/types.pb';
import PromotePipeline from './PromotePipeline';
import PromotionInfo from './PromotionInfo';
import Target from './Target';
import { EnvironmentCard } from './styles';

const getStrategy = (promo?: Promotion) => {
  if (!promo) return '-';
  if (!promo.manual) return 'Automated';

  const nonNullStrat = _.map(promo.strategy, (value, key) => {
    if (value !== null) return key;
  });
  return _.startCase(nonNullStrat[0] || '-');
};

const EnvironmentContainer = styled(Flex)`
  background: ${props => props.theme.colors.pipelineGray}};
  border-radius: 8px;
  padding: ${props => props.theme.spacing.small};
`;

const PromotionContainer = styled.div`
  height: 40px;
  padding: ${props => props.theme.spacing.small} 0;
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
        const strategy = env.promotion
          ? getStrategy(env.promotion)
          : getStrategy(pipeline.promotion);

        return (
          <EnvironmentContainer key={index} tall column gap="16">
            <PromotionInfo targets={status} />
            <EnvironmentCard background={index}>
              <Flex column gap="8" wide>
                <Flex between align wide>
                  <Text bold capitalize size="large">
                    {env.name}
                  </Text>
                  <Text>{env.targets?.length || '0'} TARGETS</Text>
                </Flex>
                <Flex gap="8" wide start>
                  <Text bold>Strategy:</Text>
                  <Text> {strategy}</Text>
                </Flex>
              </Flex>
              <PromotionContainer>
                {strategy !== 'Automated' &&
                  index < environments.length - 1 && (
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
`;
