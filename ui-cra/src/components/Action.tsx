import React, { FC, MouseEventHandler } from 'react';
import styled, { css } from 'styled-components';
import theme from 'weaveworks-ui-components/lib/theme';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { IconDefinition } from '@fortawesome/fontawesome-svg-core';
import { faExternalLinkAlt } from '@fortawesome/free-solid-svg-icons';
import { SafeAnchor } from './Shared';
import { GitOpsBlue } from './../muiTheme';

const actionStyle = css`
  display: flex;
  align-items: center;
  font-size: ${theme.fontSizes.normal};
  white-space: nowrap;

  &:enabled {
    color: ${GitOpsBlue};
  }

  &:hover:enabled {
    color: ${theme.colors.blue700};
  }
`;

const ActionAnchor = styled(SafeAnchor)`
  ${actionStyle}
`;

const ActionButton = styled.button`
  ${actionStyle}

  background: transparent;
  border: 2px solid rgba(224, 224, 224, 1);
  outline: 0;
  padding: ${theme.spacing.xs} ${theme.spacing.small};

  &:enabled {
    color: ${props =>
      props.className === 'danger' ? theme.colors.orange600 : GitOpsBlue};
  }
  &:hover:enabled {
    cursor: pointer;
    background-color: ${theme.colors.gray50};
    color: ${props =>
      props.className === 'danger'
        ? theme.colors.orange700
        : theme.colors.blue700};
  }
  &:hover:disabled {
    pointer-events: none;
  }
`;

export const ActionIcon = styled(FontAwesomeIcon)`
  margin-right: ${theme.spacing.xs};
`;

interface BaseAction {
  icon?: IconDefinition;
  text: string;
}

interface OnClickActionProps extends BaseAction {
  onClick: (event: MouseEventHandler<HTMLButtonElement>) => void;
}

export const OnClickAction: FC<
  OnClickActionProps & React.HTMLProps<HTMLButtonElement>
> = ({ id, onClick, icon, text, className, disabled }) => (
  <ActionButton
    id={id}
    className={className}
    onClick={onClick}
    disabled={disabled}
  >
    <ActionIcon icon={icon ?? faExternalLinkAlt} />
    {text}
  </ActionButton>
);

interface AnchorActionProps extends BaseAction {
  href: string;
}

export const AnchorAction: FC<AnchorActionProps> = ({ href, icon, text }) => (
  <ActionAnchor href={href}>
    <ActionIcon icon={icon ?? faExternalLinkAlt} />
    {text}
  </ActionAnchor>
);
