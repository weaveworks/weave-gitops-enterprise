import { createStyles, makeStyles } from '@material-ui/styles';
import { Link, theme } from '@weaveworks/weave-gitops';
import styled from 'styled-components';

const { small, xs, medium, base, large, xxl } = theme.spacing;

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

export const OnBoardingMessageWrapper = styled.div`
  background: rgba(255, 255, 255, 0.85);
  box-shadow: 5px 10px 50px 3px rgb(0 0 0 / 10%);
  border-radius: 10px;
  padding: ${large} ${xxl};
  max-width: 560px;
  margin: auto;
`;

export const Header4 = styled.div`
  font-size: ${theme.fontSizes.large};
  font-weight: 600;
  color: ${theme.colors.neutral30};
  margin-bottom: ${small};
`;

export const TextWrapper = styled.p`
  font-size: ${theme.fontSizes.medium};
  color: ${theme.colors.neutral30};
  font-weight: 400;
`;

export const LinkTag = styled(Link)`
  color: ${theme.colors.primary};
`;
