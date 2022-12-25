import { Theme } from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import { theme } from '@weaveworks/weave-gitops';
const { xxs } = theme.spacing;
const {
  successOriginal,
  alertOriginal,
} = theme.colors;
export const useSecretStyle = makeStyles(() =>
  createStyles({


  }),
);
