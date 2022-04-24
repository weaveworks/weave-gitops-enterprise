import { Theme } from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import { theme } from '@weaveworks/weave-gitops';

export const usePolicyStyle= makeStyles((wtheme: Theme) =>
    createStyles({
      contentWrapper: {
        margin: `${theme.spacing.small} 0`,
      },
      sectionSeperator: {
        margin: `${theme.spacing.medium} 0`,
      },
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
        padding: `${theme.spacing.xxs} ${theme.spacing.xs}`,
        marginLeft: theme.spacing.xs,
        fontWeight: 400,
        fontSize: theme.fontSizes.tiny,
      },
      codeWrapper: {
        background:  theme.colors.neutral10,
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
        marginTop: theme.spacing.xs,
        background: theme.colors.neutral10,
        padding: theme.spacing.xs,
        maxHeight: '300px',
        fontSize: theme.fontSizes.tiny,
        whiteSpace: 'pre-wrap',
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
      code: {
        wordBreak: 'break-word',
      },
    }),
  );
