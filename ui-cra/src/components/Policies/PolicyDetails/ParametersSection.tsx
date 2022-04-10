import React from 'react';
import { Policy } from '../../../capi-server/capi_server.pb';
import { createStyles, makeStyles } from '@material-ui/styles';

const useStyles = makeStyles(() =>
  createStyles({
    cardTitle: {
      fontWeight: 700,
      fontSize: '14px',
      color: '#737373',
      marginBottom: '12px',
    },
    body1: {
      fontWeight: 400,
      fontSize: '14px',
      color: '#1A1A1A',
      marginLeft: '8px',
    },
    labelText: {
      fontWeight: 400,
      fontSize: '12px',
    },
    parameterWrapper: {
      border: '1px solid #DADDE0',
      boxSizing: 'border-box',
      borderRadius: '2px',
      padding: '16px',
      display: 'flex',
    },
    parameterInfo: {
      display: 'flex',
      alignItems: 'start',
      flexDirection: 'column',
    },
  }),
);

function ParametersSection({ parameters }: Policy) {
  const classes = useStyles();
  return (
    <>
      <div>
        <span className={classes.cardTitle}>Parameters Definition</span>
        {parameters?.map(parameter => (
          <div className={classes.parameterWrapper}>
            <div className={classes.parameterInfo}>
              <span className={classes.labelText}>Parameter Name</span>
              <span className={classes.body1}>{parameter.name}</span>
            </div>
            <div className={classes.parameterInfo}>
              <span className={classes.labelText}>Parameter Type</span>
              <span className={classes.body1}>{parameter.type}</span>
            </div>
            <div className={classes.parameterInfo}>
              <span className={classes.labelText}>Value</span>
              <span className={classes.body1}>{parameter.value.value}</span>
            </div>
            <div className={classes.parameterInfo}>
              <span className={classes.labelText}>Required</span>
              <span className={classes.body1}>{parameter.required}</span>
            </div>
          </div>
        ))}
      </div>
    </>
  );
}

export default ParametersSection;
