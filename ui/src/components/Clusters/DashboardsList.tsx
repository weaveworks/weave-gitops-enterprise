import useClusters from '../../hooks/clusters';
import { GitopsClusterEnriched } from '../../types/custom';
import { List, ListItem } from '@material-ui/core';
import { Link } from '@weaveworks/weave-gitops';
import { FC } from 'react';
import styled from 'styled-components';

// FIXME: move "a" styling up to a top level CSS rule
const ListWrapper = styled(List)`
  padding: 0;
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
    <ListWrapper>
      {Object.entries(annotations).map(([key, value]) => (
        <ListItem key={key}>
          {
            <Link href={value} newTab>
              {key}
            </Link>
          }
        </ListItem>
      ))}
    </ListWrapper>
  ) : (
    <>-</>
  );
};
