import { Flex } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import styled from 'styled-components';
import { PipelineTargetStatus } from '../../../api/pipelines/types.pb';
import { EnvironmentCard } from './styles';

type Props = {
  className?: string;
  targets: PipelineTargetStatus[];
};

const PromotionInfoContainer = styled(EnvironmentCard)`
  background: #f6f7f9;
  border: 1px solid black;
  border-style: dashed;
`;
export enum ReadyType {
  Ready = 'Ready',
  NotReady = 'Not Ready',
  Reconciling = 'Reconciling',
  PendingAction = 'PendingAction',
  Suspended = 'Suspended',
  None = 'None',
}

export enum ReadyStatusValue {
  True = 'True',
  False = 'False',
  Unknown = 'Unknown',
  None = 'None',
}

export function computeReady(conditions: any): ReadyType | undefined {
  if (!conditions?.length) return undefined;
  const readyCondition = _.find(
    conditions,
    c => c.type === 'Ready' || c.type === 'Available',
  );

  if (readyCondition) {
    if (readyCondition.status === ReadyStatusValue.True) {
      return ReadyType.Ready;
    }

    if (readyCondition.status === ReadyStatusValue.Unknown) {
      if (readyCondition.reason === 'Progressing') return ReadyType.Reconciling;
      if (readyCondition.reason === 'TerraformPlannedWithChanges')
        return ReadyType.PendingAction;
    }

    if (readyCondition.status === ReadyStatusValue.None) return ReadyType.None;

    return ReadyType.NotReady;
  }

  if (_.find(conditions, c => c.status === ReadyStatusValue.False)) {
    return ReadyType.NotReady;
  }

  return ReadyType.Ready;
}
const gatherTargetStatus = (targets: PipelineTargetStatus[]) => {
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
  return `${reduced['true']} out of ${targets.length} targets ready for promotion`;
};

function PromotionInfo({ className, targets }: Props) {
  return (
    <PromotionInfoContainer className={className}>
      <Flex center align wide tall wrap>
        {gatherTargetStatus(targets)}
      </Flex>
    </PromotionInfoContainer>
  );
}

export default styled(PromotionInfo).attrs({ className: PromotionInfo.name })``;
