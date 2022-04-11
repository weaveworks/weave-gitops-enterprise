import React from 'react';
import { Policy, PolicyParam } from '../../../capi-server/capi_server.pb';
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
    chip: {
      background: 'rgba(10, 57, 64, 0.06)',
      borderRadius: '4px',
      padding: '2px 8px',
      fontWeight: 400,
      fontSize: '14px',
    },
  }),
);

// function parameterParseValue({ value, type }: PolicyParam) : React.FC<any> {
//   const classes = useStyles();

//   return (
//     <>
//      { !!value && <div className={classes.chip}>undefined</div> }
//      {/* {parseValue(type)} */}

//     </>
//     );

// }

function ParametersSection({ parameters }: Policy) {
  const classes = useStyles();
  const parseValue = (parameter: PolicyParam) => {
    switch (parameter.type) {
      case 'boolean':
        return parameter.value.value ? 'true' : 'false';
      case 'array':
        return parameter.value.value.join(', ');
      case 'string':
        return parameter.value.value;
      case 'integer':
        return parameter.value.value.toString();
    }
  };
  return (
    <>
      <div>
        <span className={classes.cardTitle}>Parameters Definition</span>
        {parameters?.map((parameter: PolicyParam) => (
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
                  parseValue(parameter)
                ) : (
                  <div className={classes.chip}>undefined</div>
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
