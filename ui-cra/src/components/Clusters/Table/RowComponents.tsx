import GitUrlParse from 'git-url-parse';
import { groupBy, orderBy, sortBy } from 'lodash';
import React, { FC } from 'react';
import styled from 'styled-components';
import { theme } from '@weaveworks/weave-gitops';
import { Cluster, Node } from '../../../types/kubernetes';
import { intersperse } from '../../../utils/formatters';
import { NameLink, NotAvailable, SafeAnchor, Tooltip } from '../../Shared';

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
  color: ${theme.colors.neutral40};
`;

const ClusterBitsContainer = styled.span`
  color: hsl(0, 0%, 60%);
`;

const ClusterNodesTooltipCell = styled.td`
  padding: 2px 4px;
`;

const WorkspacesTooltipCell = styled.td`
  padding: 2px 4px;
`;

const MoreWorkspacesRow = styled.td`
  padding-top: 9px;
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

export const ClusterNodeVersions: FC<{ cluster: Cluster }> = ({ cluster }) => {
  if (!cluster.nodes) {
    return <NotAvailable>Not available</NotAvailable>;
  }
  const sortedVersionGroups = sortBy(
    groupBy(cluster.nodes, 'kubeletVersion'),
    (node, version) => version,
  );

  return (
    <Tooltip title={<NodesTooltip nodes={cluster.nodes} />}>
      <span>
        {sortedVersionGroups.map(nodes => (
          <span key={nodes[0].kubeletVersion}>
            {nodes[0].kubeletVersion} <ClusterNodes nodes={nodes} />
          </span>
        ))}
      </span>
    </Tooltip>
  );
};

const WorkspacesContainer = styled.div`
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: no-wrap;
  min-width: 100%;
  width: 200px;
`;

export const WorkspacesLink: FC<{ cluster: Cluster; wsNames: string }> = ({
  cluster,
  wsNames,
}) => {
  const { ingressUrl } = cluster;

  // join the urls properly taking into account potentially repeated `/`s.
  const workspacesUrl = ingressUrl && new URL('/workspaces', ingressUrl).href;

  return ingressUrl ? (
    <NameLink href={workspacesUrl}>{wsNames}</NameLink>
  ) : (
    <span>{wsNames}</span>
  );
};

export const Workspaces: FC<{ cluster: Cluster }> = React.forwardRef(
  ({ cluster, ...props }, ref) => {
    if (!cluster.workspaces) {
      return <NotAvailable>Not available</NotAvailable>;
    }
    if (cluster.workspaces.length === 0) {
      return <NotAvailable>0 workspaces</NotAvailable>;
    }

    const wsNames: string[] = cluster.workspaces.map(
      workspace => workspace.name,
    );
    return (
      <WorkspacesContainer
        ref={ref as React.RefObject<HTMLDivElement> | null | undefined}
        {...props}
      >
        <WorkspacesLink cluster={cluster} wsNames={wsNames.join(', ')} />
      </WorkspacesContainer>
    );
  },
);

export const WorkspacesTooltip: FC<{ workspaces: string[] }> = ({
  workspaces,
}) => {
  let wsList: string[];
  let wsNumber = 0;
  const numWorkspaces = 10;
  if (workspaces.length > numWorkspaces) {
    wsList = workspaces.slice(0, numWorkspaces);
    wsNumber = workspaces.length - numWorkspaces;
  } else {
    wsList = workspaces;
  }

  return (
    <table>
      <tbody>
        {wsList.map(ws => (
          <tr key={ws}>
            <WorkspacesTooltipCell>{ws}</WorkspacesTooltipCell>
          </tr>
        ))}
        {wsNumber ? (
          <>
            <tr>
              <MoreWorkspacesRow key="spacer" />
            </tr>
            <tr key="more-ws">
              <WorkspacesTooltipCell>
                {' '}
                {wsNumber} more...{' '}
              </WorkspacesTooltipCell>
            </tr>
          </>
        ) : null}
      </tbody>
    </table>
  );
};
