import { CheckCircle, Error, RemoveCircle } from '@material-ui/icons';
import LinearProgress from '@material-ui/core/LinearProgress';

import { useCanaryStyle } from '../CanaryStyles';

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
  Ready = 'Succeeded',
}


function CanaryStatus({
  status,
  canaryWeight,
}: {
  status: string;
  canaryWeight: number;
}) {
  const classes = useCanaryStyle();

  return (
    <div className={classes.statusWrapper}>
      {(() => {
        switch (status) {
          case CanaryDeploymentStatus.Waiting:
          case CanaryDeploymentStatus.WaitingPromotion:
            return (
              <>
                <RemoveCircle className={`${classes.statusWaiting}`} />
                {status}
              </>
            );
          case CanaryDeploymentStatus.Succeeded:
          case CanaryDeploymentStatus.Initialized:
            return (
              <>
                <CheckCircle className={`${classes.statusReady}`} />
                {status}
              </>
            );
          case CanaryDeploymentStatus.Initializing:
            return (
              <>
                <LinearProgress
                  variant="determinate"
                  value={50}
                  classes={{
                    barColorPrimary: classes.barroot,
                  }}
                  className={classes.root}
                />
                <span> {Math.ceil(canaryWeight / 100)} / 10</span>
              </>
            );
          case CanaryDeploymentStatus.Failed:
            return (
              <>
                <Error className={`${classes.statusFailed}`} />
                {status}
              </>
            );
          default:
            return <>{status}</>;
        }
      })()}
    </div>
  );
}

export default CanaryStatus;
