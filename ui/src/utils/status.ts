import { IconType } from '@weaveworks/weave-gitops';

enum ObjectStatus {
  Success = 'Success',
  Failed = 'Failed',
  Reconciling = 'Reconciling',
  Suspended = 'Suspended',
  PendingAction = 'PendingAction',
  NoStatus = '-'
}

type IndicatorInfo = {
  color: string;
  type: IconType;
};

export const getIndicatorInfo = (
  status: string | undefined,
): IndicatorInfo => {
  switch (status) {
    case ObjectStatus.Success:
      return {
        color: 'successOriginal',
        type: IconType.SuccessIcon,
      };
    case ObjectStatus.Reconciling:
      return {
        color: 'primary',
        type: IconType.ReconcileIcon,
      };
    case ObjectStatus.Suspended:
      return {
        color: 'feedbackOriginal',
        type: IconType.SuspendedIcon,
      };
    case ObjectStatus.PendingAction:
      return {
        color: 'feedbackOriginal',
        type: IconType.PendingActionIcon,
      };
    case ObjectStatus.NoStatus:
      return {
        color: 'neutral20',
        type: IconType.RemoveCircleIcon,
      }; 
    default:
      return {
        color: 'alertOriginal',
        type: IconType.ErrorIcon,
      };
  }
};