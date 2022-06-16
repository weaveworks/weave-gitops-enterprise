import { CheckCircle, Error, RemoveCircle } from '@material-ui/icons';
import LinearProgress from '@material-ui/core/LinearProgress';

import styled from 'styled-components';
import { useCanaryStyle } from '../CanaryStyles';

enum CanaryDeploymentStatus {
  Waiting = 'Initializing',
  Ready = 'Succeeded',
  Progressing = 'Progressing',
  Failed = 'Failed',
}

const BorderLinearProgress = styled(LinearProgress)(() => ({
  height: 10,
  width: '100%',
  marginRight: '8px',
  borderRadius: 5,
  [`&.barColorPrimary`]: {
    backgroundColor: '#D8D8D8',
  },
  [`&.bar`]: {
    borderRadius: 5,
    backgroundColor: '#1a90ff',
  },
}));

function CanaryStatus({
  status = '',
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
            return (
              <>
                <RemoveCircle className={`${classes.statusWaiting}`} />
                Waiting
              </>
            );
          case CanaryDeploymentStatus.Ready:
            return (
              <>
                <CheckCircle className={`${classes.statusReady}`} />
                Ready
              </>
            );
          case CanaryDeploymentStatus.Progressing:
            return (
              <>
                <BorderLinearProgress
                  variant="determinate"
                  value={canaryWeight}
                />
                <span> {Math.ceil(canaryWeight / 100)} / 10</span>
              </>
            );
          default:
            return (
              <>
                <Error className={`${classes.statusFailed}`} />
                Failed
              </>
            );
        }
      })()}
    </div>
  );
}

export default CanaryStatus;
