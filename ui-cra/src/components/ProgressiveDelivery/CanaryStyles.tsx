import { createStyles, makeStyles } from '@material-ui/styles';

export const useCanaryStyle = makeStyles(() =>
  createStyles({
    rowHeaderWrapper: {
      margin: `${12} 0`,
      display: 'flex',
      justifyContent: 'start',
      alignItems: 'center',
    },
    cardTitle: {
      fontWeight: 600,
      fontSize: 14,
    },
    body1: {
      fontWeight: 400,
      fontSize: 14,
      color: '#1a1a1a',
      marginLeft: 8,
      display: 'flex',
      justifyContent: 'start',
      alignItems: 'center',
    },
    colorGreen: {
      color: '#27AE60',
    },
    statusWrapper: {
      display: 'flex',
      gap: 8,
      justifyContent: 'start',
      alignItems: 'center',
    },
    statusMessage: {
      color: '#9E9E9E', // add natural25 to core
    },
    statusReady: {
      color: '#27AE60',
    },
    statusWaiting: {
      color: '#F2994A',
    },
    statusFailed: {
      color: '#BC3B1D',
    },
    sectionHeaderWrapper: {
      background: '#F6F7F9',
      padding: `16 8`,
      margin: `16 0`,
    },
    straegyIcon: {
      marginLeft: 12,
    },
    barroot: {
      backgroundColor: '#27AE60',
    },
    statusProcessing: {
      backgroundColor: '#d8d8d8',
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
      padding: 12,
      marginTop: 24,
      cursor: 'pointer',
    },
    expandableSpacing: {
      marginLeft: 8,
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
