import React, { FC } from 'react';
import { isEmpty } from 'lodash';
import styled from 'styled-components';
import { transparentize } from 'polished';
import { makeStyles, createStyles } from '@material-ui/core/styles';
import { Flex, theme as weaveTheme } from '@weaveworks/weave-gitops';
import ArrowForward from '@material-ui/icons/ArrowForwardIos';
import { Link } from 'react-router-dom';

interface Size {
  size?: 'small';
}

const Span = styled.span`
  color: ${({ theme }) => theme.colors.white};
  text-overflow: ellipsis;
  white-space: nowrap;
  overflow: hidden;
  max-width: 300px;
`;
export const Title = styled.div<Size>`
  margin-right: ${({ size }) =>
    size === 'small' ? weaveTheme.spacing.xxs : weaveTheme.spacing.xs};
  white-space: nowrap;
`;
export const Count = styled.div<Size>`
  background: ${({ size }) =>
    size === 'small'
      ? transparentize(0.5, weaveTheme.colors.primaryLight10)
      : weaveTheme.colors.primaryLight10};
  padding: 0 ${weaveTheme.spacing.xxs};
  align-self: center;
  font-size: ${({ size }) =>
    size === 'small'
      ? weaveTheme.fontSizes.small
      : weaveTheme.fontSizes.medium};
  color: ${weaveTheme.colors.black};
  margin-left: ${weaveTheme.spacing.xxs};
  border-radius: ${weaveTheme.borderRadius.soft};
`;
export interface Breadcrumb {
  label: string;
  url?: string;
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
      padding: ` 0 ${weaveTheme.spacing.small}`,
    },
    iconDivider: {
      fontSize: '16px',
    },
    link: {
      color: weaveTheme.colors.black,
      fontWeight: 400,
    },
    labelLink: {
      color: weaveTheme.colors.black,
      fontWeight: 700,
      marginRight: weaveTheme.spacing.xs,
    },
  }),
);
export const Breadcrumbs: FC<Props> = ({ path, size }) => {
  const classes = useStyles();
  return (
    <Flex
      style={{
        fontSize: '20px',
        height: '32px',
      }}
    >
      {path.map(({ label, url }, index) => (
        <div
          key={index}
          className={classes.path}
          style={{ alignItems: 'center' }}
        >
          {index > 0 && (
            <div className={classes.divider}>
              <ArrowForward className={classes.iconDivider} />
            </div>
          )}
          {isEmpty(url) ? (
            <Span className={classes.labelLink} title={label}>
              {label}
            </Span>
          ) : (
            <Title role="heading" size={size}>
              <Link to={url || ''} className={classes.link}>
                {label}
              </Link>
            </Title>
          )}
        </div>
      ))}
    </Flex>
  );
};
