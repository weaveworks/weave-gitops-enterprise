import React, { FC } from 'react';
import { List, ListItem } from '@material-ui/core';
import { GitopsClusterEnriched } from '../../types/custom';
import styled from 'styled-components';
import useClusters from './../../contexts/Clusters';

const ListWrapper = styled(List)`
  td[class*='MuiTableCell-root'] {
    padding: 0;
  },
  overflow: elipsis,
`;

export const DashboardsList: FC<{
  cluster: GitopsClusterEnriched;
}> = ({ cluster }) => {
  const { getDashboardAnnotations } = useClusters();
  const annotations = getDashboardAnnotations(cluster);

  return (
    <ListWrapper>
      {Object.entries(annotations).map(([key, value]) => {
        return (
          <ListItem key={key} disableGutters>
            <a href={value}>{key}</a>
          </ListItem>
        );
      })}
    </ListWrapper>
  );
};
