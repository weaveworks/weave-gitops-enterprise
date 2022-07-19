import React, { FC } from 'react';
import styled from 'styled-components';
import useClusters from './../../contexts/Clusters';
import { List, ListItem } from '@material-ui/core';
import { GitopsClusterEnriched } from '../../types/custom';
import * as DOMPurify from 'dompurify';

// FIXME: move "a" styling up to a top level CSS rule
const ListWrapper = styled(List)`
  li[class*='MuiListItem-root'] {
    padding: 0 0 ${props => props.theme.spacing.xxs} 0;
  }
  a {
    color: ${({ theme }) => theme.colors.primary};
  }
`;

//
// This should clean "javascript:"" and "data:" schemes which can both be
// used to craft malicious links:
// - "javascript:alert(1);"
// - "data:text/html,<script>alert(document.domain)</script>"
//
// It leaves relative links in tact (like "/flux_runtime")
//
// DOMPurify has a slightly awkward API but is supposedly one of the more
// battle-tested sanitization libraries.
//
const cleanHref = (href: string): string | undefined => {
  const a = document.createElement('a');
  a.href = href;
  const cleanAnchor = DOMPurify.sanitize(a, { RETURN_DOM: true });
  // return undefined as "href: string | undefined"
  return cleanAnchor?.querySelector('a')?.href || undefined;
};

export const DashboardsList: FC<{
  cluster: GitopsClusterEnriched;
}> = ({ cluster }) => {
  const { getDashboardAnnotations } = useClusters();
  const annotations = getDashboardAnnotations(cluster);

  return Object.keys(annotations).length > 0 ? (
    <ListWrapper style={{ padding: 0 }}>
      {Object.entries(annotations).map(([key, value]) => (
        <ListItem key={key}>
          <a href={cleanHref(value)} target="_blank" rel="noopener noreferrer">
            {key}
          </a>
        </ListItem>
      ))}
    </ListWrapper>
  ) : (
    <>-</>
  );
};
