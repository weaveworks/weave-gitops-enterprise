import { createStyles, makeStyles } from '@material-ui/styles';
import { theme } from '@weaveworks/weave-gitops';

const { small, xs, medium, base } = theme.spacing;

export const useCanaryStyle = makeStyles(() =>
  createStyles({
    rowHeaderWrapper: {
      margin: `${small} 0`,
      display: 'flex',
      justifyContent: 'start',
      alignItems: 'center',
    },
    cardTitle: {
      fontWeight: 600,
      fontSize: theme.fontSizes.medium,
    },
    body1: {
      fontWeight: 400,
      fontSize: theme.fontSizes.medium,
      color: theme.colors.black,
      marginLeft: xs,
      display: 'flex',
      justifyContent: 'start',
      alignItems: 'center',
    },
    colorGreen: {
      color: theme.colors.successOriginal,
    },
    statusWrapper: {
      display: 'flex',
      gap: xs,
      justifyContent: 'start',
      alignItems: 'center',
    },
    statusMessage: {
      color: '#9E9E9E', // add natural25 to core
    },
    statusReady: {
      color: theme.colors.successOriginal,
    },
    statusWaiting: {
      color: '#F2994A',
    },
    statusFailed: {
      color: theme.colors.alertOriginal,
    },
    sectionHeaderWrapper: {
      background: theme.colors.neutralGray,
      padding: `${base} ${xs}`,
      margin: `${base} 0`,
    },
    straegyIcon: {
      marginLeft: small,
    },
    barroot: {
      backgroundColor: theme.colors.successOriginal,
    },
    statusProcessing: {
      backgroundColor: theme.colors.neutral20,
      width: '100%',
      height: 8,
      borderRadius: 5,
      minWidth: '75px',
    },
    statusProcessingText: {
      minWidth: 'fit-content',
    },
    code: {
      wordBreak: 'break-word',
    },
    expandableCondition: {
      display: 'flex',
      justifyContent: 'start',
      alignItems: 'center',
      padding: small,
      marginTop: medium,
      cursor: 'pointer',
    },
    expandableSpacing: {
      marginLeft: xs,
    },
    fadeIn: {
      transform: 'scaleY(0)',
      transformOrigin: 'top',
      display: 'block',
      maxHeight: 0,
      transition: 'transform 0.15s ease',
    },
    fadeOut: {
      transform: 'scaleY(1)',
      transformOrigin: 'top',
      transition: 'transform 0.15s ease',
    },
  }),
);
