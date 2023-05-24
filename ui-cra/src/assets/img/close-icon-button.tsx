import { IconButton, IconButtonProps } from '@material-ui/core';
import { Close } from '@material-ui/icons';
import styled from 'styled-components';

const CloseIconButton = ({ onClick }: IconButtonProps) => (
  <IconButton onClick={onClick}>
    <Close />
  </IconButton>
);

export default styled(CloseIconButton).attrs({
  className: CloseIconButton.name,
})`
  .MuiIconButton-root {
    position: absolute;
    right: ${props => props.theme.spacing.xs};
    top: ${props => props.theme.spacing.xs};
    color: ${props => props.theme.colors.neutral20};
  }
`;
