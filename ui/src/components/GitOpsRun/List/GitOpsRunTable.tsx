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
import { Routes } from '../../../utils/nav';
import CommandCell from './CommandCell';

const sessionObjectsInfo = 'session objects created';

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
  const name = kind === Kind.Kustomization ? 'run-dev-ks' : 'run-dev-helm';
  const clusterName = `${metadata.annotations['run.weave.works/namespace']}/${metadata.name}`;
  const route =
    kind === Kind.Kustomization ? V2Routes.Kustomization : V2Routes.HelmRelease;

  const text = `${kind}/${name}`;

  return s.info === sessionObjectsInfo ? (
    <Link
      to={formatURL(route, {
        name,
        namespace: 'flux-system',
        clusterName,
      })}
    >
      {text}
    </Link>
  ) : (
    <>{text}</>
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
    <DataTable
      key={sessions?.length}
      filters={initialFilterState}
      rows={sessions}
      fields={[
        {
          label: 'Name',
          value: (s: FluxObject) => {
            const namespace =
              s.obj.metadata.annotations['run.weave.works/namespace'];
            return (
              <Link
                to={formatURL(Routes.GitOpsRunDetail, {
                  name: s.name,
                  namespace,
                })}
              >
                {s.name}
              </Link>
            );
          },
          sortValue: ({ name }: FluxObject) => name,
          textSearchable: true,
        },

        {
          label: 'Automation',
          value: s => <AutomationLink s={s} />,
        },
        {
          label: 'Source',
          value: s => {
            const metadata = s.obj.metadata;
            const clusterName = `${metadata.annotations['run.weave.works/namespace']}/${metadata.name}`;
            const text = 'Bucket/run-dev-bucket';

            return s.info === sessionObjectsInfo ? (
              <Link
                to={formatURL(V2Routes.Bucket, {
                  name: 'run-dev-bucket',
                  namespace: 'flux-system',
                  clusterName,
                })}
              >
                {text}
              </Link>
            ) : (
              text
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
            if (ports?.length) {
                return <PortLinks ports={ports} />;
            }
            return null;
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
            <Timestamp
              tooltip
              time={obj.metadata.creationTimestamp}
              hideSeconds
            />
          ),
          sortValue: ({ obj }) => obj.metadata.creationTimestamp,
          minWidth: 175,
        },
      ]}
    />
  );
};

export default GitOpsRunTable;
