import LinearProgress from '@material-ui/core/LinearProgress';

import { Flex, Icon, IconType } from '@weaveworks/weave-gitops';
import styled from 'styled-components';

enum CanaryDeploymentStatus {
  Initializing = 'Initializing',
  Initialized = 'Initialized',
  Waiting = 'Waiting',
  Progressing = 'Progressing',
  WaitingPromotion = 'WaitingPromotion',
  Promoting = 'Promoting',
  Finalising = 'Finalising',
  Succeeded = 'Succeeded',
  Failed = 'Failed',
  Terminating = 'Terminating',
  Terminated = 'Terminated',
  Ready = 'Ready',
}
const MuiLinearProgress = styled(LinearProgress)`
  background-color: ${props => props.theme.colors.neutral20};
  width: 100%;
  height: 8px;
  border-radius: 5px;
  min-width: 75px;
  .barroot {
    background-color: ${props => props.theme.colors.successOriginal};
    border-radius: 5px;
  }
`;
function CanaryStatus({
  status,
  value = { current: 0, total: 0 },
}: {
  status: string;
  value?: { current: number; total: number };
}) {
  return (
    <Flex gap="8" align start>
      {(() => {
        switch (status) {
          case CanaryDeploymentStatus.Waiting:
          case CanaryDeploymentStatus.WaitingPromotion:
            return (
              <>
                <Icon
                  type={IconType.RemoveCircleIcon}
                  color="alertOriginal"
                  size="medium"
                />
                {status}
              </>
            );
          case CanaryDeploymentStatus.Succeeded:
          case CanaryDeploymentStatus.Initialized:
          case CanaryDeploymentStatus.Ready:
            return (
              <>
                <Icon
                  type={IconType.CheckCircleIcon}
                  color="successOriginal"
                  size="medium"
                />
                {status}
              </>
            );
          case CanaryDeploymentStatus.Progressing:
            const current =
              value.current > value.total ? value.total : value.current;
            return (
              <>
                <MuiLinearProgress
                  variant="determinate"
                  value={(current / value.total) * 100}
                  classes={{
                    barColorPrimary: 'barroot',
                  }}
                />
                <span
                  style={{ minWidth: 'fit-content' }}
                >{`${current} / ${value.total}`}</span>
              </>
            );
          case CanaryDeploymentStatus.Failed:
            return (
              <>
                <Icon
                  type={IconType.SuspendedIcon}
                  color="feedbackOriginal"
                  size="medium"
                />
                {status}
              </>
            );
          default:
            return <>{status}</>;
        }
      })()}
    </Flex>
  );
}

export default CanaryStatus;
