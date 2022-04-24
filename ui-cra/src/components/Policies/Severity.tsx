import { CallReceived, CallMade, Remove } from '@material-ui/icons';
import { usePolicyStyle } from './PolicyStyles';

function Severity({ severity }: { severity: string }) {
  const classes = usePolicyStyle();
  switch (severity.toLocaleLowerCase()) {
    case 'low':
      return (
        <div className={classes.flexStart}>
          <CallReceived
            className={`${classes.severityIcon} ${classes.severityLow}`}
          />
          <span className={classes.capitlize}>{severity}</span>
        </div>
      );
    case 'high':
      return (
        <div className={classes.flexStart}>
          <CallMade
            className={`${classes.severityIcon} ${classes.severityHigh}`}
          />
          <span className={classes.capitlize}>{severity}</span>
        </div>
      );
    case 'medium':
      return (
        <div className={classes.flexStart}>
          <Remove
            className={`${classes.severityIcon} ${classes.severityMedium}`}
          />
          <span className={classes.capitlize}>{severity}</span>
        </div>
      );

    default:
      return <span className={classes.capitlize}>{severity}</span>;
  }
}

export default Severity;
