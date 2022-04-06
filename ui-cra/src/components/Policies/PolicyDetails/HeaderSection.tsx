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
    chip: {
      background: 'rgba(10, 57, 64, 0.06)',
      borderRadius: '2px',
      padding: '2px 8px',
      marginLeft: '8px',
      fontWeight: 400,
      fontSize: '11px',
    },
  }),
);

function HeaderSection({ id, tags, severity, category, targets }: Policy) {
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
      <div>
        <span className={classes.cardTitle}>Severity:</span>
        <span className={classes.body1}>{severity}</span>
      </div>
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
    </>
  );
}

export default HeaderSection;
