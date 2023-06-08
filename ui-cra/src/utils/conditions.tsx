import _ from 'lodash';

export interface Condition {
  type: string;
  status: string;
  reason: string;
  message: string;
  timestamp: string;
}

export enum ReadyType {
  Uknown = 'Unknown',
  Ready = 'Ready',
  NotReady = 'Not Ready',
  Reconciling = 'Reconciling',
}

export function computeReady(conditions: Condition[]): ReadyType {
  if (!conditions) {
    return ReadyType.Uknown;
  }

  const readyCondition =
    _.find(conditions, c => c.type === 'Ready') ||
    _.find(conditions, c => c.type === 'Available');
  if (readyCondition) {
    if (readyCondition.status === 'True') return ReadyType.Ready;
    if (
      readyCondition.status === 'Unknown' &&
      readyCondition.reason === 'Progressing'
    )
      return ReadyType.Reconciling;
  }
  return ReadyType.NotReady;
}

export function computeMessage(conditions: Condition[]): string {
  const readyCondition =
    _.find(conditions, c => c.type === 'Ready') ||
    _.find(conditions, c => c.type === 'Available');

  return readyCondition ? readyCondition.message : 'unknown error';
}
