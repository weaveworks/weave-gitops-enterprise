import { CheckCircle, Error } from '@material-ui/icons';
import { useSecretStyle } from './SecretStyles';

function Status({ status }: { status: string }) {
  const classes = useSecretStyle();
  switch (status.toLocaleLowerCase()) {
    case 'ready':
      return (
        <div className={classes.flexStart}>
          <CheckCircle
            className={`${classes.readyIcon} ${classes.statusIcon}`}
          />
          <span className={classes.capitlize}>{status}</span>
        </div>
      );
    case 'not ready':
      return (
        <div className={classes.flexStart}>
          <Error
            className={`${classes.notReadyIcon} ${classes.statusIcon}`}
          />
          <span className={classes.capitlize}>{status}</span>
        </div>
      );

    default:
      return <span className={classes.capitlize}>{status}</span>;
  }
}

export default Status;
