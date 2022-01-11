import React, { FC } from 'react';
import { IconButton, IconButtonProps } from '@material-ui/core';
import { withStyles } from '@material-ui/styles';
import { Close } from '@material-ui/icons';
import { theme } from '@weaveworks/weave-gitops';

const StyledIconButton = withStyles(() => ({
  root: {
    position: 'absolute',
    right: theme.spacing.xs,
    top: theme.spacing.xs,
    color: theme.colors.neutral20,
  },
}))(IconButton);

export const CloseIconButton: FC<IconButtonProps> = ({ onClick }) => (
  <StyledIconButton onClick={onClick}>
    <Close />
  </StyledIconButton>
);
