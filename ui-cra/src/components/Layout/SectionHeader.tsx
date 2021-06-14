import React, { FC } from 'react';
import styled from 'styled-components';
import theme from 'weaveworks-ui-components/lib/theme';
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
  background: #00b3ec;
  height: ${64}px;
  flex-grow: 1;
  padding-left: ${theme.spacing.large};
  position: sticky;
  top: 0;
  z-index: 2;
`;

interface Props extends Size {
  icon?: JSX.Element | null;
  className?: string;
  path: Breadcrumb[];
}

export const SectionHeader: FC<Props> = ({
  children,
  size,
  className,
  path,
}) => (
  <Wrapper className={className} size={size}>
    <Breadcrumbs path={path} />
    {children}
  </Wrapper>
);
