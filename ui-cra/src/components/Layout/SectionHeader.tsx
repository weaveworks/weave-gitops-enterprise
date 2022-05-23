import React, { FC } from 'react';
import styled from 'styled-components';
import { theme, UserSettings } from '@weaveworks/weave-gitops';
import { Breadcrumb, Breadcrumbs } from '../Breadcrumbs';

interface Size {
  size?: 'small';
}

const Wrapper = styled.div<Size>`
  align-items: center;
  justify-content: space-between;
  display: flex;
  color: ${({ size }) =>
    size === 'small' ? theme.colors.neutral40 : 'inherit'};
  font-size: ${({ size }) => (size === 'small' ? 16 : 20)}px;
  background: ${theme.colors.primary};
  height: ${80}px;
  flex-grow: 1;
  padding-left: ${theme.spacing.large};
  position: sticky;
  top: 0;
  z-index: 2;

  .MuiListItemIcon-root {
    min-width: 30px;
  }
`;

interface Props extends Size {
  icon?: JSX.Element | null;
  className?: string;
  path?: Breadcrumb[];
}

export const SectionHeader: FC<Props> = ({
  children,
  size,
  className,
  path,
}) => {
  return (
    <Wrapper className={className} size={size}>
      {path ? <Breadcrumbs path={path} /> : null}
      {children}
      <UserSettings />
    </Wrapper>
  );
};
