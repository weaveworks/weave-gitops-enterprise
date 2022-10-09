import React, { FC } from 'react';
import styled from 'styled-components';
import useClusters from '../../hooks/clusters';
import { List, ListItem } from '@material-ui/core';
import { GitopsClusterEnriched } from '../../types/custom';
import { isAllowedLink, Link } from '@weaveworks/weave-gitops';

// FIXME: move "a" styling up to a top level CSS rule
const ListWrapper = styled(List)`
  li[class*='MuiListItem-root'] {
    padding: 0 0 ${props => props.theme.spacing.xxs} 0;
  }
  a {
    color: ${({ theme }) => theme.colors.primary};
  }
`;

export const DashboardsList: FC<{
  cluster: GitopsClusterEnriched;
}> = ({ cluster }) => {
  const { getDashboardAnnotations } = useClusters();
  const annotations = getDashboardAnnotations(cluster);

  return Object.keys(annotations).length > 0 ? (
    <ListWrapper style={{ padding: 0 }}>
      {Object.entries(annotations).map(([key, value]) => (
        <ListItem key={key}>
          {isAllowedLink(value) ? (
            <Link href={value} newTab>
              {key}
            </Link>
          ) : (
            <span>{key}</span>
          )}
        </ListItem>
      ))}
    </ListWrapper>
  ) : (
    <>-</>
  );
};
