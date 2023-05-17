import { Flex, UserSettings } from '@weaveworks/weave-gitops';
import { FC } from 'react';
import styled from 'styled-components';
import { Breadcrumb, Breadcrumbs } from '../Breadcrumbs';

interface Size {
  size?: 'small';
}

const Wrapper = styled(Flex)<Size>`
  align-items: center;
  justify-content: space-between;
  color: ${props =>
    props.size === 'small' ? props.theme.colors.neutral40 : 'inherit'};
  font-size: ${({ size }) => (size === 'small' ? 16 : 20)}px;
  height: ${60}px;
  flex-grow: 1;
  padding: ${props =>
    `0 ${props.theme.spacing.small} 0 ${props.theme.spacing.medium}`};
  position: sticky;
  top: 0;
  z-index: 2;
  background: inherit;
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
