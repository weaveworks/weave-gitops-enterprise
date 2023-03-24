import { Theme } from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import { Flex, theme } from '@weaveworks/weave-gitops';
import { Link } from 'react-router-dom';
import styled from 'styled-components';

const { xxs, xs, small, medium, base, none } = theme.spacing;
const {
  neutral10,
  neutral20,
  neutral30,
  black,
  primary,
  primary20,
  feedbackDark,
  alertDark,
} = theme.colors;

export const usePolicyStyle = makeStyles((wtheme: Theme) =>
  createStyles({
    contentWrapper: {
      margin: `${small} 0`,
    },
    sectionSeperator: {
      margin: `calc(${medium}*2) 0`,
    },
    cardTitle: {
      fontWeight: 700,
      fontSize: theme.fontSizes.medium,
      color: neutral30,
    },
    body1: {
      fontWeight: 400,
      fontSize: theme.fontSizes.medium,
      color: black,
      marginTop: xs,
    },
    chip: {
      background: neutral10,
      borderRadius: xxs,
      padding: `${xxs} ${xs}`,
      marginLeft: xs,
    },
    codeWrapper: {
      background: neutral10,
      borderRadius: xxs,
      padding: `${small} ${base}`,
      marginLeft: none,
    },

    marginrightSmall: {
      marginRight: xs,
    },
    marginTopSmall: {
      marginTop: xs,
    },
    editor: {
      '& a': {
        color: primary,
      },
      '& > *:first-child': {
        marginTop: 0,
      },
      '& > *:last-child': {
        marginBottom: 0,
      },
      marginTop: xs,
      background: neutral10,
      padding: small,
      maxHeight: '300px',
      overflow: 'scroll',
    },
    severityIcon: {
      fontSize: theme.fontSizes.large,
      marginRight: xxs,
    },
    severityLow: {
      color: primary20,
    },
    severityMedium: {
      color: feedbackDark,
    },
    severityHigh: {
      color: alertDark,
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
      color: neutral30,
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
      borderBottom: `1px solid ${neutral20}`,
    },
    normalRow: {
      borderBottom: `1px solid ${neutral20}`,
    },
    normalCell: {
      padding: wtheme.spacing(2),
    },
    link: {
      color: primary,
      fontWeight: 600,
      whiteSpace: 'pre-line',
    },
    canaryLink: {
      color: primary,
      fontWeight: 600,
      display: 'flex',
      justifyContent: 'start',
      alignItems: 'center',
    },
    code: {
      wordBreak: 'break-word',
    },
    titleNotification: {
      color: primary,
    },
    occurrencesList: {
      paddingLeft: wtheme.spacing(1),
      marginLeft: xs,
    },
    messageWrape: {
      whiteSpace: 'normal',
    },
    labelText: {
      fontWeight: 400,
      fontSize: theme.fontSizes.medium,
      color: neutral30,
    },
    inlineFlex: {
      display: 'inline-flex',
      marginRight: xs,
    },
    modeIcon: {
      fontSize: theme.fontSizes.large,
      marginRight: xxs,
      color: neutral30,
    },
  }),
);

export const ParameterWrapper = styled(Flex)`
  border: 1px solid ${neutral20};
  box-sizing: border-box;
  border-radius: ${xxs};
  padding: ${base};
  margin-bottom: ${base};
  margin-top: ${base};
`;

export const ParameterInfo = styled(Flex)`
  width: 100%;
  font-weight: 400;
  font-size: ${theme.fontSizes.medium};
  .label {
    color: ${neutral30};
  }
  .body1 {
    color: black;
    margin-top: ${xs};
  }
`;

export const LinkWrapper = styled(Link)`
  color: ${primary};
  font-weight: 600;
`;

export const ChipWrapper = styled.span`
  background: ${neutral10};
  border-radius: ${xs};
  padding: ${xxs} ${xs};
  margin-right: ${xs};
`;

export const ModeWrapper = styled.div`
  align-items: center;
  justify-content: flex-start;
  display: inline-flex;
  margin-right: ${xs};
  svg {
    color: ${neutral30};
    font-size: 20px;
    margin-right: 4px;
  }
  span {
    text-transform: capitalize;
  }
`;
