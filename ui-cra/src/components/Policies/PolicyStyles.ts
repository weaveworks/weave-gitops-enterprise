import { Theme } from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import { theme } from '@weaveworks/weave-gitops';

export const usePolicyStyle = makeStyles((wtheme: Theme) =>
  createStyles({
    contentWrapper: {
      margin: `${theme.spacing.small} 0`,
    },
    sectionSeperator: {
      margin: `calc(${theme.spacing.medium}*2) 0`,
    },
    cardTitle: {
      fontWeight: 700,
      fontSize: theme.fontSizes.normal,
      color: theme.colors.neutral30,
    },
    body1: {
      fontWeight: 400,
      fontSize: theme.fontSizes.normal,
      color: theme.colors.black,
      marginLeft: theme.spacing.xs,
    },
    chip: {
      background: theme.colors.neutral10,
      borderRadius: theme.spacing.xxs,
      padding: `${theme.spacing.xxs} ${theme.spacing.xs}`,
      marginLeft: theme.spacing.xs,
    },
    codeWrapper: {
      background: theme.colors.neutral10,
      borderRadius: theme.spacing.xxs,
      padding: `${theme.spacing.small} ${theme.spacing.base}`,
      marginLeft: theme.spacing.none,
    },

    marginrightSmall: {
      marginRight: theme.spacing.xs,
    },
    marginTopSmall: {
      marginTop: theme.spacing.xs,
    },
    editor: {
      '& a': {
        color: theme.colors.primary,
      },
      '& > *:first-child': {
        marginTop: 0,
      },
      '& > *:last-child': {
        marginBottom: 0,
      },
      marginTop: theme.spacing.xs,
      background: theme.colors.neutral10,
      padding: theme.spacing.small,
      maxHeight: '300px',
      overflow: 'scroll',
    },
    severityIcon: {
      fontSize: theme.fontSizes.small,
      marginRight: theme.spacing.xxs,
    },
    severityLow: {
      color: '#006B8E',
    },
    severityMedium: {
      color: '#8A460A',
    },
    severityHigh: {
      color: '#9F3119',
    },
    column: {
      flexDirection: 'column',
    },
    flexStart: {
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'flex-start',
    },
    capitlize: {
      textTransform: 'capitalize',
    },
    headerCell: {
      color: theme.colors.neutral30,
      fontWeight: 700,
    },
    paper: {
      marginBottom: 10,
      marginTop: 10,
      overflowX: 'auto',
      width: '100%',
    },
    root: {
      width: '100%',
    },
    table: {
      whiteSpace: 'nowrap',
    },
    tableHead: {
      borderBottom: `1px solid ${theme.colors.neutral20}`,
    },

    normalRow: {
      borderBottom: `1px solid ${theme.colors.neutral20}`,
    },
    normalCell: {
      padding: wtheme.spacing(2),
    },
    link: {
      color: theme.colors.primary,
      fontWeight: 600,
      whiteSpace: 'pre-line',
    },
    canaryLink: {
      color: theme.colors.primary,
      fontWeight: 600,
      display: 'flex',
      justifyContent: 'start',
      alignItems: 'center',
    },
    code: {
      wordBreak: 'break-word',
    },
    titleNotification: {
      color: theme.colors.primary,
    },
    occurrencesList: {
      paddingLeft: wtheme.spacing(1),
    },
    messageWrape: {
      whiteSpace: 'normal',
    },
    labelText: {
      fontWeight: 400,
      fontSize: theme.fontSizes.tiny,
      color: theme.colors.neutral30,
    },
    parameterWrapper: {
      border: `1px solid ${theme.colors.neutral20}`,
      boxSizing: 'border-box',
      borderRadius: theme.spacing.xxs,
      padding: theme.spacing.base,
      display: 'flex',
      marginBottom: theme.spacing.base,
      marginTop: theme.spacing.base,
    },
    parameterInfo: {
      display: 'flex',
      alignItems: 'start',
      flexDirection: 'column',
      width: '100%',
    },
  }),
);
