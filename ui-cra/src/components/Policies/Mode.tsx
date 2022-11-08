import { VerifiedUser, Policy } from '@material-ui/icons';
import { usePolicyStyle } from './PolicyStyles';

function Mode({ modeName }: { modeName: string }) {
  const classes = usePolicyStyle();
  return (
    <>
      {modeName.toLocaleLowerCase() === 'audit' ? (
        <div className={`${classes.flexStart} ${classes.inlineFlex}`}>
          <VerifiedUser className={classes.modeIcon} />
          <span className={classes.capitlize}>{modeName}</span>
        </div>
      ) : (
        modeName.toLocaleLowerCase() === 'admission' && (
          <div className={`${classes.flexStart} ${classes.inlineFlex}`}>
            <Policy className={classes.modeIcon} />
            <span className={classes.capitlize}>Enforce</span>
          </div>
        )
      )}
    </>
  );
}

export default Mode;
