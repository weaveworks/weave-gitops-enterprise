import React, { FC } from 'react';
import styled from 'styled-components';
import { theme } from '@weaveworks/weave-gitops';
import useClusters from './../../contexts/Clusters';
import { List, ListItem } from '@material-ui/core';
import { GitopsClusterEnriched } from '../../types/custom';

const ListWrapper = styled(List)`
  li[class*='MuiListItem-root'] {
    padding: 0 0 4px 0;
  }
  a {
    color: ${theme.colors.primary};
  }
`;

export const DashboardsList: FC<{
  cluster: GitopsClusterEnriched;
}> = ({ cluster }) => {
  const { getDashboardAnnotations } = useClusters();
  const annotations = getDashboardAnnotations(cluster);

  return (
    <ListWrapper style={{ padding: 0 }}>
      {Object.entries(annotations).map(([key, value]) => {
        return (
          <ListItem key={key}>
            <a href={value} target="_blank" rel="noopener noreferrer">
              {key}
            </a>
          </ListItem>
        );
      })}
    </ListWrapper>
  );
};
