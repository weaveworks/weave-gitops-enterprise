import React, { FC, ReactElement } from 'react';
import styled, { css } from 'styled-components';
import theme from 'weaveworks-ui-components/lib/theme';
import { darken } from 'polished';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { IconDefinition } from '@fortawesome/fontawesome-svg-core';
import { faExternalLinkAlt } from '@fortawesome/free-solid-svg-icons';
import { SafeAnchor } from '../Shared';
import { Breadcrumb, Breadcrumbs } from '../Breadcrumbs';

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
  actions?: SectionHeaderAction[] | null;
  icon?: JSX.Element | null;
  className?: string;
  path: Breadcrumb[];
}

export const SectionHeader: FC<Props> = ({
  actions,
  children,
  size,
  className,
  path,
}) => (
  <Wrapper className={className} size={size}>
    <Breadcrumbs path={path} />
    {actions?.map(action => action)}
    {children}
  </Wrapper>
);
