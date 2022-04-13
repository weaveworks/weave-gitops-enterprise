import React from 'react';
import { Policy } from '../../../capi-server/capi_server.pb';
import { createStyles, makeStyles } from '@material-ui/styles';
import Severity from '../Severity';
import styled from 'styled-components';
import { theme } from '@weaveworks/weave-gitops';

const useStyles = makeStyles(() =>
  createStyles({
    cardTitle: {
      fontWeight: 700,
      fontSize: theme.fontSizes.small,
      color: theme.colors.neutral30,
    },
    body1: {
      fontWeight: 400,
      fontSize: theme.fontSizes.small,
      color: theme.colors.black,
      marginLeft: theme.spacing.xs,
    },
    chip: {
      background: theme.colors.neutral10,
      borderRadius: theme.spacing.xxs,
      padding: `${theme.spacing.xs} ${theme.spacing.xxs}`,
      marginLeft: theme.spacing.xs,
      fontWeight: 400,
      fontSize: theme.fontSizes.tiny,
    },
    codeWrapper: {
      background: theme.colors.neutral10,
      borderRadius: theme.spacing.xxs,
      padding: `${theme.spacing.small} ${theme.spacing.base}`,
      marginLeft: theme.spacing.none,
    },
    paddingTopSmall: {
      paddingTop: theme.spacing.xs,
    },
    marginrightSmall: {
      marginRight: theme.spacing.xs,
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

  console.log(theme.spacing);
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
