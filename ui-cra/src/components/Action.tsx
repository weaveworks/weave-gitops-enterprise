import React, { FC, MouseEventHandler } from 'react';
import styled, { css } from 'styled-components';
import { theme } from '@weaveworks/weave-gitops';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { IconDefinition } from '@fortawesome/fontawesome-svg-core';
import { faExternalLinkAlt } from '@fortawesome/free-solid-svg-icons';
import { SafeAnchor } from './Shared';

const actionStyle = css`
  display: flex;
  align-items: center;
  font-size: ${theme.fontSizes.normal};
  white-space: nowrap;

  &:enabled {
    color: ${theme.colors.primary};
  }

  &:hover:enabled {
    color: ${theme.colors.primaryDark};
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
      props.className === 'danger' ? theme.colors.alert : theme.colors.primary};
  }
  &:hover:enabled {
    cursor: pointer;
    background-color: ${theme.colors.neutral20};
    color: ${props =>
      props.className === 'danger'
        ? theme.colors.alert
        : theme.colors.primaryDark};
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
