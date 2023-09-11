import _ from 'lodash';

export interface Condition {
  type: string;
  status: string;
  reason: string;
  message: string;
  timestamp: string;
}

export function computeMessage(conditions: Condition[]): string {
  const readyCondition =
    _.find(conditions, c => c.type === 'Ready') ||
    _.find(conditions, c => c.type === 'Available');

  return readyCondition ? readyCondition.message : 'unknown error';
}
