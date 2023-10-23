import { Flex, Text, computeReady } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { PipelineTargetStatus } from '../../../api/pipelines/types.pb';
import { EnvironmentCard } from './styles';

type Props = {
  className?: string;
  targets: PipelineTargetStatus[];
};

const PromotionInfoContainer = styled(EnvironmentCard)`
  background: ${props => props.theme.colors.neutralGray};
  border: 1px solid ${props => props.theme.colors.neutral40};
  border-style: dashed;
`;

const gatherTargetStatus = (targets: PipelineTargetStatus[]) => {
  if (!targets.length) return 'No targets found.';
  const reduced = targets.reduce(
    (obj, target) => {
      let ready = true;
      target.workloads?.forEach(workload => {
        if (computeReady(workload.conditions || []) !== 'Ready') ready = false;
      });
      if (ready) obj.true += 1;
      else obj.false += 1;
      return obj;
    },
    { true: 0, false: 0 },
  );
  return `${reduced['true']} out of ${targets.length} targets have successfully updated.`;
};

function PromotionInfo({ className, targets }: Props) {
  return (
    <PromotionInfoContainer className={className}>
      <Flex center align wide tall wrap>
        <Text color="neutral40">{gatherTargetStatus(targets)}</Text>
      </Flex>
    </PromotionInfoContainer>
  );
}

export default styled(PromotionInfo).attrs({ className: PromotionInfo.name })`
  ${Text} {
    text-align: center;
  }
`;
