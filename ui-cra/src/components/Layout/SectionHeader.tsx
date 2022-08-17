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
  height: ${80}px;
  flex-grow: 1;
  padding-left: ${theme.spacing.large};
  position: sticky;
  top: 0;
  z-index: 2;

  background: inherit;
  .MuiListItemIcon-root {
    min-width: 30px;
  }
`;

// This should be removed once CORE change the icon in UserSettings
const CustomUserSettings = styled(UserSettings)`
  svg {
    fill: ${({ theme }) => theme.colors.primary} !important;
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
      <CustomUserSettings />
    </Wrapper>
  );
};
