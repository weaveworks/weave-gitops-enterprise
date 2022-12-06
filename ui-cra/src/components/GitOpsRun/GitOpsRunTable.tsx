import {
  DataTable,
  filterConfig,
  Flex,
  formatURL,
  Kind,
  Link,
  Timestamp,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { FluxObject } from '@weaveworks/weave-gitops/ui/lib/objects';
import { FC } from 'react';
import { TableWrapper } from '../Shared';
import CommandCell from './CommandCell';

const PortLinks: React.FC<{ ports: string }> = ({ ports = '' }) => {
  const list = ports.split(',');
  return (
    <Flex column>
      {list.map(port => (
        <Link key={port} href={`http://localhost:${port}`} newTab>
          http://localhost:{port}
        </Link>
      ))}
    </Flex>
  );
};

const AutomationLink: React.FC<{ s: FluxObject }> = ({ s }) => {
  const metadata = s.obj.metadata;
  const kind =
    metadata.annotations['run.weave.works/automation-kind'] === 'ks'
      ? Kind.Kustomization
      : Kind.HelmRelease;
  const namespace =
    metadata.namespace === 'default' ? 'flux-system' : metadata.namespace;
  const name = kind === Kind.Kustomization ? 'run-dev-ks' : 'run-dev-helm';
  const route =
    kind === Kind.Kustomization ? V2Routes.Kustomization : V2Routes.HelmRelease;

  return (
    <Link
      to={formatURL(route, {
        name,
        namespace: namespace,
        clusterName: `${metadata.namespace}/${metadata.name}`,
      })}
    >
      {kind}/{name}
    </Link>
  );
};
interface Props {
  sessions: FluxObject[];
}

const GitOpsRunTable: FC<Props> = ({ sessions }) => {
  let initialFilterState = {
    ...filterConfig(
      sessions,
      'CLI Version',
      session =>
        session.obj.metadata.annotations['run.weave.works/cli-version'],
    ),
    ...filterConfig(
      sessions,
      'Port Forward',
      session =>
        session.obj.metadata.annotations['run.weave.works/port-forward'],
    ),
  };

  return (
    <TableWrapper id="gitopsRun-list">
      <DataTable
        key={sessions?.length}
        filters={initialFilterState}
        rows={sessions}
        fields={[
          { label: 'Name', value: 'name', textSearchable: true },
          {
            label: 'Automation',
            value: s => <AutomationLink s={s} />,
          },
          {
            label: 'Source',
            value: s => {
              const metadata = s.obj.metadata;
              return (
                <Link
                  to={formatURL(V2Routes.Bucket, {
                    name: 'run-dev-bucket',
                    namespace: 'flux-system',
                    clusterName: `${metadata.namespace}/${metadata.name}`,
                  })}
                >
                  Bucket/run-dev-bucket
                </Link>
              );
            },
          },
          {
            label: 'CLI Version',
            value: ({ obj }) =>
              obj.metadata.annotations['run.weave.works/cli-version'],
          },
          {
            label: 'Port Forward',
            value: ({ obj }) => {
              const ports: string =
                obj.metadata.annotations['run.weave.works/port-forward'];
              return <PortLinks ports={ports} />;
            },
            sortValue: ({ obj }) =>
              obj.metadata.annotations['run.weave.works/port-forward'],
          },
          {
            label: 'Command',
            value: ({ obj }) => (
              <CommandCell
                command={obj.metadata.annotations['run.weave.works/command']}
              />
            ),
            sortValue: ({ obj }) =>
              obj.metadata.annotations['run.weave.works/command'],
          },
          {
            label: 'Created',
            value: ({ obj }) => (
              <Timestamp time={obj.metadata.creationTimestamp} hideSeconds />
            ),
            sortValue: ({ obj }) => obj.metadata.creationTimestamp,
            minWidth: 175,
          },
        ]}
      />
    </TableWrapper>
  );
};

export default GitOpsRunTable;
