import React, { FC } from 'react';
import { IconButton, IconButtonProps, Theme } from '@material-ui/core';
import { withStyles } from '@material-ui/styles';
import { Close } from '@material-ui/icons';

const StyledIconButton = withStyles((theme: Theme) => ({
  root: {
    position: 'absolute',
    right: theme.spacing(1),
    top: theme.spacing(1),
    color: theme.palette.grey[500],
  },
}))(IconButton);

export const CloseIconButton: FC<IconButtonProps> = ({ onClick }) => (
  <StyledIconButton onClick={onClick}>
    <Close />
  </StyledIconButton>
);
