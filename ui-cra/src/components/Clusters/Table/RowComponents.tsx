import GitUrlParse from 'git-url-parse';
import { groupBy, orderBy } from 'lodash';
import React, { FC } from 'react';
import styled from 'styled-components';
import { theme } from '@weaveworks/weave-gitops';
import { Node } from '../../../types/kubernetes';
import { intersperse } from '../../../utils/formatters';
import { SafeAnchor } from '../../Shared';

const getGitRepoHTTPSURL = (repoUrl?: string, repoBranch?: string): string => {
  if (repoUrl) {
    const parsedRepo = GitUrlParse(repoUrl);
    return `https://${parsedRepo.resource}/${parsedRepo.full_name}/tree/${repoBranch}`;
  }
  return '';
};

const Link = styled(SafeAnchor)`
  color: ${theme.colors.primary};
  &:hover {
    color: ${theme.colors.primaryDark};
  }
`;

export const RepoLink: FC<{ url: string; branch: string }> = ({
  url,
  branch,
}) => <Link href={getGitRepoHTTPSURL(url, branch)}>View git repo</Link>;

const ClusterBit = styled.span`
  color: ${theme.colors.neutral30};
`;

const ClusterBitsContainer = styled.span`
  color: ${theme.colors.neutral30};
`;

const ClusterNodesTooltipCell = styled.td`
  padding: 2px 4px;
`;

export const ClusterNodes: FC<{ nodes: Node[] }> = ({ nodes }) => {
  const { true: controlPlanes, false: workers } = groupBy(
    nodes,
    'isControlPlane',
  );
  const controlPlaneBits = controlPlanes
    ? [<ClusterBit key="controlPlanes">{controlPlanes.length}CP</ClusterBit>]
    : [];
  const workerBits = workers
    ? [<ClusterBit key="workers">{workers.length}</ClusterBit>]
    : [];
  return (
    <ClusterBitsContainer>
      ({' '}
      {intersperse([...controlPlaneBits, ...workerBits], index => (
        <span key={`sep-number-${index}`}> | </span>
      ))}{' '}
      )
    </ClusterBitsContainer>
  );
};

export const NodesTooltip: FC<{ nodes: Node[] }> = ({ nodes }) => {
  const tooltipRows = orderBy(
    groupBy(nodes, node => `${node.isControlPlane}-${node.kubeletVersion}`),
    ['0.isControlPlane', '0.kubeletVersion'],
    ['desc', 'asc'],
  );

  return (
    <table>
      <tbody>
        {tooltipRows.map(nodes => (
          <tr key={String(nodes[0].isControlPlane)}>
            <ClusterNodesTooltipCell>
              {nodes.length}{' '}
              {nodes[0].isControlPlane ? 'Control plane nodes' : 'worker nodes'}
            </ClusterNodesTooltipCell>
            <ClusterNodesTooltipCell>
              {nodes[0].kubeletVersion}
            </ClusterNodesTooltipCell>
          </tr>
        ))}
      </tbody>
    </table>
  );
};
