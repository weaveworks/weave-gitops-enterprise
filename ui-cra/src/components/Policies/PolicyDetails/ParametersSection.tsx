import React from 'react';
import { Policy } from '../../../capi-server/capi_server.pb';
import { createStyles, makeStyles } from '@material-ui/styles';
import styled from 'styled-components';

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
    },
    labelText: {
      fontWeight: 400,
      fontSize: '12px',
      color: '#737373',
    },
    parameterWrapper: {
      border: '1px solid #DADDE0',
      boxSizing: 'border-box',
      borderRadius: '2px',
      padding: '16px',
      display: 'flex',
      marginBottom: '16px',
      marginTop: '16px',
    },
    parameterInfo: {
      display: 'flex',
      alignItems: 'start',
      flexDirection: 'column',
      width: '100%',
    },
  }),
);
const Chip = styled.div`
  background-color: #f8fafa;
  border-radius: 4px;
  padding: 4px;
`;

function ParametersSection({ parameters }: Policy) {
  const classes = useStyles();
  return (
    <>
      <div>
        <span className={classes.cardTitle}>Parameters Definition</span>
        {parameters?.map(parameter => (
          <div key={parameter.name} className={classes.parameterWrapper}>
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
              <span className={classes.body1}>
                {parameter.value ? (
                  parameter.type === 'array' ? (
                    parameter.value.value.join(', ')
                  ) : (
                    parameter.value.value
                  )
                ) : (
                  <Chip>undefined</Chip>
                )}
              </span>
            </div>
            <div className={classes.parameterInfo}>
              <span className={classes.labelText}>Required</span>
              <span className={classes.body1}>
                {parameter.required ? 'True' : 'False'}
              </span>
            </div>
          </div>
        ))}
      </div>
    </>
  );
}

export default ParametersSection;
