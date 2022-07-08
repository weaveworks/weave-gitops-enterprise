import React, { FC } from 'react';
import styled from 'styled-components';
import useClusters from './../../contexts/Clusters';
import { List, ListItem } from '@material-ui/core';
import { GitopsClusterEnriched } from '../../types/custom';

const ListWrapper = styled(List)`
  li[class*='MuiListItem-root'] {
    padding: 0 0 ${props => props.theme.spacing.xxs} 0;
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
          <a href={value} target="_blank" rel="noopener noreferrer">
            {key}
          </a>
        </ListItem>
      ))}
    </ListWrapper>
  ) : (
    <>-</>
  );
};
