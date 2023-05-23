import { createStyles, makeStyles } from '@material-ui/core/styles';
import ArrowForward from '@material-ui/icons/ArrowForwardIos';
import { Flex, Link } from '@weaveworks/weave-gitops';
import { isEmpty } from 'lodash';
import { transparentize } from 'polished';
import { FC } from 'react';
import styled from 'styled-components';

interface Size {
  size?: 'small';
}

const Container = styled(Flex)`
  font-size: ${20}px;
  height: 32px;
`;

const Span = styled.span`
  text-overflow: ellipsis;
  white-space: nowrap;
  overflow: hidden;
  max-width: 300px;
  .labelLink {
    color: ${props => props.theme.colors.black};
    fontWeight: 700;
    marginRight: ${props => props.theme.spacing.xs};
  },
`;

const Divider = styled.div`
  display: flex;
  alignitems: center;
  justifycontent: center;
  padding: 0 ${props => props.theme.spacing.small};
`;

export const Title = styled.div<Size>`
  margin-right: ${props =>
    props.size === 'small' ? props.theme.spacing.xxs : props.theme.spacing.xs};
  white-space: nowrap;
`;

export const Count = styled.div<Size>`
  background: ${props =>
    props.size === 'small'
      ? transparentize(0.5, props.theme.colors.primaryLight10)
      : props.theme.colors.primaryLight10};
  padding: 0 ${props => props.theme.spacing.xxs};
  align-self: center;
  font-size: ${props =>
    props.size === 'small'
      ? props.theme.fontSizes.small
      : props.theme.fontSizes.medium};
  color: ${props => props.theme.colors.black};
  margin-left: ${props => props.theme.spacing.xxs};
  border-radius: ${props => props.theme.borderRadius.soft};
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
    iconDivider: {
      fontSize: '16px',
    },
  }),
);

export const Breadcrumbs: FC<Props> = ({ path, size }) => {
  const classes = useStyles();
  return (
    <Container align center className="test-id-breadcrumbs">
      {path.map(({ label, url }, index) => (
        <div
          key={index}
          className={classes.path}
          style={{ alignItems: 'center' }}
        >
          {index > 0 && (
            <Divider>
              <ArrowForward className={classes.iconDivider} />
            </Divider>
          )}
          {isEmpty(url) ? (
            <Span className="labelLink" title={label}>
              {label}
            </Span>
          ) : (
            <Title role="heading" size={size}>
              <Link to={url || ''} textProps={{ color: 'black' }}>
                {label}
              </Link>
            </Title>
          )}
        </div>
      ))}
    </Container>
  );
};
