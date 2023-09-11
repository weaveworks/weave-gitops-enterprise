import { IconButton, IconButtonProps } from '@material-ui/core';
import { Close } from '@material-ui/icons';
import styled from 'styled-components';

const CloseIconButton = ({ onClick, className }: IconButtonProps) => (
  <IconButton onClick={onClick} className={className}>
    <Close />
  </IconButton>
);

export default styled(CloseIconButton).attrs({
  className: CloseIconButton.name,
})`
  &.MuiIconButton-root {
    color: ${props => props.theme.colors.black};
  }
`;
