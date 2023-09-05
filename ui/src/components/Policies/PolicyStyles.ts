import { Theme } from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import { Link } from 'react-router-dom';
import styled from 'styled-components';

export const usePolicyStyle = makeStyles((wtheme: Theme) =>
  createStyles({
    root: {
      width: '100%',
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
  }),
);

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
