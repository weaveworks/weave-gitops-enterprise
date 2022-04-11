import React from 'react';
import { Policy } from '../../../capi-server/capi_server.pb';
import { createStyles, makeStyles } from '@material-ui/styles';
import Severity from '../Severity';
import styled from 'styled-components';

const useStyles = makeStyles(() =>
  createStyles({
    cardTitle: {
      fontWeight: 700,
      fontSize: '14px',
      color: '#737373',
    },
    body1: {
      fontWeight: 400,
      fontSize: '14px',
      color: '#1A1A1A',
      marginLeft: '8px',
    },
    chip: {
      background: 'rgba(10, 57, 64, 0.06)',
      borderRadius: '4px',
      padding: '2px 8px',
      marginLeft: '8px',
      fontWeight: 400,
      fontSize: '11px',
    },
    codeWrapper: {
      background: '#F8FAFA',
      borderRadius: '4px',
      padding: '10px 16px',
      marginLeft: 0,
    },
    paddingTopSmall: {
      paddingTop: '8px',
    },
    marginrightSmall: {
      marginRight: '8px',
    },
    flexStart: {
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'start',
    },
  }),
);
const FlexStart = styled.div`
  display: flex;
  align-items: center;
  justify-content: start;
`;

function HeaderSection({
  id,
  tags,
  severity,
  category,
  targets,
  description,
  howToSolve,
  code,
}: Policy) {
  const classes = useStyles();
  return (
    <>
      <div>
        <span className={classes.cardTitle}>Policy ID:</span>
        <span className={classes.body1}>{id}</span>
      </div>
      <div>
        <span className={classes.cardTitle}>Tags:</span>
        {tags?.map(tag => (
          <span key={tag} className={classes.chip}>
            {tag}
          </span>
        ))}
      </div>
      <FlexStart>
        <span className={`${classes.cardTitle} ${classes.marginrightSmall}`}>
          Severity:
        </span>
        <Severity severity={severity || ''} />
      </FlexStart>

      <div>
        <span className={classes.cardTitle}>Category:</span>
        <span className={classes.body1}>{category}</span>
      </div>
      <div>
        <span className={classes.cardTitle}>Targeted K8s Kind:</span>
        {targets?.kinds?.map(kind => (
          <span key={kind} className={classes.chip}>
            {kind}
          </span>
        ))}
      </div>

      <hr />

      <div className={classes.paddingTopSmall}>
        <span className={classes.cardTitle}>Description</span>
        <p className={`${classes.body1} ${classes.codeWrapper}`}>
          {description}
        </p>
      </div>
      <div>
        <span className={classes.cardTitle}>How To Resolve</span>
        <p className={`${classes.body1} ${classes.codeWrapper}`}>
          {howToSolve}
        </p>
      </div>
    </>
  );
}

export default HeaderSection;
