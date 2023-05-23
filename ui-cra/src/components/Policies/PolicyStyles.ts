import { Theme } from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import { Flex } from '@weaveworks/weave-gitops';
import { Link } from 'react-router-dom';
import styled from 'styled-components';

export const usePolicyStyle = makeStyles((wtheme: Theme) =>
  createStyles({
    contentWrapper: {
      margin: `${12} 0`,
    },
    sectionSeperator: {
      margin: `calc(${24}*2) 0`,
    },
    cardTitle: {
      fontWeight: 700,
      fontSize: 14,
      color: '#737373',
    },
    body1: {
      fontWeight: 400,
      fontSize: 14,
      color: '#1a1a1a',
      marginTop: 8,
    },
    chip: {
      background: '#f5f5f5',
      borderRadius: 4,
      padding: `${4} ${8}`,
      marginLeft: 8,
    },
    codeWrapper: {
      background: '#f5f5f5',
      borderRadius: 4,
      padding: `${12} ${16}`,
      marginLeft: 0,
    },

    marginrightSmall: {
      marginRight: 8,
    },
    marginTopSmall: {
      marginTop: 8,
    },
    editor: {
      '& a': {
        color: '#00b3ec',
      },
      '& > *:first-child': {
        marginTop: 0,
      },
      '& > *:last-child': {
        marginBottom: 0,
      },
      marginTop: 8,
      background: '#f5f5f5',
      padding: 12,
      maxHeight: '300px',
      overflow: 'scroll',
    },
    severityIcon: {
      fontSize: 20,
      marginRight: 4,
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
      color: '#737373',
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
      borderBottom: `1px solid "#d8d8d8"`,
    },
    normalRow: {
      borderBottom: `1px solid "#d8d8d8"`,
    },
    normalCell: {
      padding: wtheme.spacing(2),
    },
    link: {
      color: '#00b3ec',
      fontWeight: 600,
      whiteSpace: 'pre-line',
    },
    canaryLink: {
      color: '#00b3ec',
      fontWeight: 600,
      display: 'flex',
      justifyContent: 'start',
      alignItems: 'center',
    },
    code: {
      wordBreak: 'break-word',
    },
    titleNotification: {
      color: '#00b3ec',
    },
    occurrencesList: {
      paddingLeft: wtheme.spacing(1),
      marginLeft: 8,
    },
    messageWrape: {
      whiteSpace: 'normal',
    },
    labelText: {
      fontWeight: 400,
      fontSize: 14,
      color: '#737373',
    },
    inlineFlex: {
      display: 'inline-flex',
      marginRight: 8,
    },
    modeIcon: {
      fontSize: 20,
      marginRight: 4,
      color: '#737373',
    },
  }),
);

export const ParameterWrapper = styled.div`
  border: 1px solid ${props => props.theme.colors.neutral20};
  box-sizing: border-box;
  border-radius: ${props => props.theme.spacing.xxs};
  padding: ${props => props.theme.spacing.base};
  display: flex;
  margin-bottom: ${props => props.theme.spacing.base};
  margin-top: ${props => props.theme.spacing.base};
`;

export const ParameterInfo = styled(Flex)`
  width: 100%;
  font-weight: 400;
  font-size: ${props => props.theme.fontSizes.medium};
  .label {
    color: ${props => props.theme.colors.neutral30};
  }
  .body1 {
    color: black;
    margin-top: ${props => props.theme.spacing.xs};
  }
`;

export const LinkWrapper = styled(Link)`
  color: ${props => props.theme.colors.primary};
  font-weight: 600;
`;

export const ChipWrapper = styled.span`
  background: ${props => props.theme.colors.neutral10};
  border-radius: ${props => props.theme.spacing.xs};
  padding: ${props => props.theme.spacing.xxs}
    ${props => props.theme.spacing.xs};
  margin-right: ${props => props.theme.spacing.xs};
`;

export const ModeWrapper = styled.div`
  align-items: center;
  justify-content: flex-start;
  display: inline-flex;
  margin-right: ${props => props.theme.spacing.xs};
  svg {
    color: ${props => props.theme.colors.neutral30};
    font-size: 20px;
    margin-right: 4px;
  }
  span {
    text-transform: capitalize;
  }
`;
