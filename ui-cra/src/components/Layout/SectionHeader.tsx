import React, { FC, ReactElement } from 'react';
import styled, { css } from 'styled-components';
import theme from 'weaveworks-ui-components/lib/theme';
import { darken, transparentize } from 'polished';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { IconDefinition } from '@fortawesome/fontawesome-svg-core';
import { faExternalLinkAlt } from '@fortawesome/free-solid-svg-icons';
import { SafeAnchor } from '../Shared';

interface Size {
  size?: 'small';
}

const Wrapper = styled.div<Size>`
  align-items: center;
  display: flex;
  color: ${({ size }) => (size === 'small' ? theme.colors.gray600 : 'inherit')};
  font-size: ${({ size }) => (size === 'small' ? 16 : 20)}px;
  margin: 0 0 ${({ size }) => (size === 'small' ? 0 : theme.spacing.base)} 0;
  flex-grow: 1;
`;

const actionStyle = css`
  color: ${darken(0.1, theme.colors.gray600)};
  display: flex;
  align-items: center;
  margin-left: ${theme.spacing.medium};
  font-size: ${theme.fontSizes.normal};
  white-space: nowrap;

  &:hover {
    color: ${theme.colors.blue700};
  }
`;

const ActionAnchor = styled(SafeAnchor)`
  ${actionStyle}
`;

const ActionButton = styled.button`
  ${actionStyle}

  background: transparent;
  border: 0;
  outline: 0;
  padding: 0;

  &:hover {
    cursor: pointer;
  }
`;

const Icon = styled.div`
  margin-right: ${theme.spacing.xs};
`;

const Title = styled.div<Size>`
  margin-right: ${({ size }) =>
    size === 'small' ? theme.spacing.xxs : theme.spacing.xs};
  white-space: nowrap;
`;

export const Count = styled.div<Size>`
  background: ${({ size }) =>
    size === 'small'
      ? transparentize(0.5, theme.colors.purple100)
      : theme.colors.purple100};
  padding: 4px 8px;
  align-self: center;
  font-size: ${({ size }) =>
    size === 'small' ? theme.spacing.small : theme.fontSizes.normal};
  color: ${theme.colors.gray600};
  margin-left: ${theme.spacing.xs};
  border-radius: ${theme.borderRadius.soft};
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
> = ({ id, onClick, icon, text, className }) => (
  <ActionButton id={id} className={className} onClick={onClick}>
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

type SectionHeaderAction =
  | ReactElement<typeof OnClickAction>
  | ReactElement<typeof AnchorAction>;

interface Props extends Size {
  action?: SectionHeaderAction | null;
  count?: number | null;
  icon?: JSX.Element | null;
  title: string;
  className?: string;
}

export const SectionHeader: FC<Props> = ({
  action,
  children,
  count,
  icon,
  size,
  title,
  className,
}) => (
  <Wrapper className={className} size={size}>
    <Title role="heading" size={size}>
      {title}
    </Title>
    {count !== null && <Count size={size}>{count}</Count>}
    {icon && <Icon>{icon}</Icon>}
    {action}
    {children}
  </Wrapper>
);
