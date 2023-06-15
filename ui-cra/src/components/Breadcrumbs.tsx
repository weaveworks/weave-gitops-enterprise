import { Flex, Icon, IconType, Link } from '@weaveworks/weave-gitops';
import { isEmpty } from 'lodash';
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

export const Title = styled.div`
  margin-right: ${props => props.theme.spacing.xxs};
  white-space: nowrap;
`;
export interface Breadcrumb {
  label: string;
  url?: string;
}
interface Props extends Size {
  path: Breadcrumb[];
}

export const Breadcrumbs: FC<Props> = ({ path }) => {
  return (
    <Container align center className="test-id-breadcrumbs">
      {path.map(({ label, url }, index) => (
        <Flex key={index} align>
          {index > 0 && (
            <Divider>
              <Icon
                type={IconType.NavigateNextIcon}
                size="large"
                color="neutral40"
              />
            </Divider>
          )}
          {isEmpty(url) ? (
            <Span className="labelLink" title={label}>
              {label}
            </Span>
          ) : (
            <Title role="heading">
              <Link
                to={url || ''}
                textProps={{ color: 'black', size: 'large' }}
              >
                {label}
              </Link>
            </Title>
          )}
        </Flex>
      ))}
    </Container>
  );
};