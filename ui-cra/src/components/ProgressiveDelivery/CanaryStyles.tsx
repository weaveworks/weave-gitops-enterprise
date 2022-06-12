import { createStyles, makeStyles } from '@material-ui/styles';
import { theme } from '@weaveworks/weave-gitops';

export const useCanaryStyle = makeStyles(() =>
  createStyles({
    rowHeaderWrapper: {
      margin: `${theme.spacing.small} 0`,
      display: 'flex',
      justifyContent: 'start',
      alignItems: 'center',
    },
    cardTitle: {
      fontWeight: 600,
      fontSize: theme.fontSizes.normal,
      color: theme.colors.neutral30,
    },
    body1: {
      fontWeight: 400,
      fontSize: theme.fontSizes.normal,
      color: theme.colors.black,
      marginLeft: theme.spacing.xs,
    },
    colorGreen: {
      color: theme.colors.success,
    },
    statusWrapper: {
      display: 'flex',
      gap: theme.spacing.xs,
      justifyContent: 'start',
      alignItems: 'center',
    },
    statusMessage: {
      color: '#9E9E9E', // add natural25 to core
    },
    statusReady: {
      color: theme.colors.success,
    },
    statusWaiting: {
      color: '#F2994A',
    },
    statusFailed: {
      color: theme.colors.alert,
    },
  }),
);
