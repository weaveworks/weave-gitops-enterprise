import {
  IconType,
} from '@weaveworks/weave-gitops';
import { getIndicatorInfo } from '../status';

describe('getIndicatorInfo', () => {
  const objects = [
    {
      status: 'Success',
      expectedInfo: {
        color: 'successOriginal',
        type: IconType.SuccessIcon,
      },
    },
    {
      status: 'Reconciling',
      expectedInfo: {
        color: 'primary',
        type: IconType.ReconcileIcon,
      },
    },
    {
      status: 'Suspended',
      expectedInfo: {
        color: 'feedbackOriginal',
        type: IconType.SuspendedIcon,
      },
    },
    {
      status: 'PendingAction',
      expectedInfo: {
        color: 'feedbackOriginal',
        type: IconType.PendingActionIcon,
      },
    },
    {
      status: '-',
      expectedInfo: {
        color: 'neutral20',
        type: IconType.RemoveCircleIcon,
      },
    },
    {
      status: 'Failed',
      expectedInfo: {
        color: 'alertOriginal',
        type: IconType.ErrorIcon,
      },
    },
    {
      status: undefined,
      expectedInfo: {
        color: 'alertOriginal',
        type: IconType.ErrorIcon,
      },
    }];

    it('returns correct indicator info for different object statuses', () => {
      objects.forEach(o => {
        const info = getIndicatorInfo(o.status)

        expect(info).toEqual(o.expectedInfo);
      });
    });
});
