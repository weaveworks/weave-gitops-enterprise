import { Theme } from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import { theme } from '@weaveworks/weave-gitops';
const { xxs } = theme.spacing;
const {
  successOriginal,
  alertOriginal,
} = theme.colors;
export const useSecretStyle = makeStyles((wtheme: Theme) =>
  createStyles({
    flexStart: {
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'flex-start',
    },
    capitlize: {
      textTransform: 'capitalize',
    },
    statusIcon: {
      fontSize: theme.fontSizes.large,
      marginRight: xxs,
    },
    readyIcon: {
      color: successOriginal,
    },
    notReadyIcon: {
      color: alertOriginal,
    },
  }),
);
