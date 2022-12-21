import { Theme } from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import { theme } from '@weaveworks/weave-gitops';
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
  successOriginal,
  alertOriginal,
} = theme.colors;
console.log(theme.colors);
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
