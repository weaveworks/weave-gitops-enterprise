import React, { FC } from 'react';
import { isEmpty } from 'lodash';
import styled from 'styled-components';
import { transparentize } from 'polished';
import BreadcrumbDivider from '../assets/img/breadcrumb-divider.svg';
import { makeStyles, createStyles } from '@material-ui/core/styles';
import { theme as weaveTheme } from '@weaveworks/weave-gitops';

interface Size {
  size?: 'small';
}
const Container = styled.div`
  align-items: flex-end;
  display: flex;
  justify-content: center;
  font-size: ${20}px;
  height: 32px;
`;
const Span = styled.span`
  color: ${({ theme }) => theme.colors.white};
`;
const Link = styled.a`
  color: ${({ theme }) => theme.colors.white};
`;
export const Title = styled.div<Size>`
  margin-right: ${({ size }) =>
    size === 'small' ? weaveTheme.spacing.xxs : weaveTheme.spacing.xs};
  white-space: nowrap;
`;
export const Count = styled.div<Size>`
  background: ${({ size }) =>
    size === 'small'
      ? transparentize(0.5, weaveTheme.colors.white)
      : weaveTheme.colors.white};
  padding: ${weaveTheme.spacing.xxs} ${weaveTheme.spacing.xs};
  align-self: center;
  font-size: ${({ size }) =>
    size === 'small' ? weaveTheme.spacing.small : weaveTheme.fontSizes.normal};
  color: ${weaveTheme.colors.primary};
  margin-left: ${weaveTheme.spacing.xxs};
  border-radius: ${weaveTheme.borderRadius.soft};
`;
export interface Breadcrumb {
  label: string;
  url?: string;
  count?: number | null;
}
interface Props extends Size {
  path: Breadcrumb[];
}
const useStyles = makeStyles(() =>
  createStyles({
    path: {
      display: 'flex',
    },
    divider: {
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      paddingLeft: weaveTheme.spacing.small,
      paddingRight: weaveTheme.spacing.small,
    },
  }),
);
export const Breadcrumbs: FC<Props> = ({ path, size }) => {
  const classes = useStyles();
  return (
    <Container>
      {path.map(({ label, url, count }, index) => (
        <div key={index} className={classes.path}>
          {index > 0 && (
            <div className={classes.divider}>
              <BreadcrumbDivider />
            </div>
          )}
          {isEmpty(url) ? (
            <Span>{label}</Span>
          ) : (
            <>
              <Title role="heading" size={size}>
                <Link href={url}>{label}</Link>
              </Title>
              <Count className="section-header-count" size={size}>
                {count || 0}
              </Count>
            </>
          )}
        </div>
      ))}
    </Container>
  );
};
