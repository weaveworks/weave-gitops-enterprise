import React, { FC } from 'react';
import styled, { css } from 'styled-components';
import theme from 'weaveworks-ui-components/lib/theme';
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
    color: #00b3ec;
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

  &:hover:enabled {
    cursor: pointer;
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
  onClick: VoidFunction;
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
