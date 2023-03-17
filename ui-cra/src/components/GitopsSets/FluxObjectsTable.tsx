import {
  DataTable,
  FluxObject,
  formatURL,
  Kind,
  KubeStatusIndicator,
  Link,
  statusSortHelper,
  useLinkResolver,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { DetailViewProps } from '@weaveworks/weave-gitops/ui/components/DetailModal';
import _ from 'lodash';
import * as React from 'react';
import styled from 'styled-components';

function objectTypeToRoute(t: Kind): V2Routes {
  switch (t) {
    case Kind.GitRepository:
      return V2Routes.GitRepo;

    case Kind.Bucket:
      return V2Routes.Bucket;

    case Kind.HelmRepository:
      return V2Routes.HelmRepo;

    case Kind.HelmChart:
      return V2Routes.HelmChart;

    case Kind.Kustomization:
      return V2Routes.Kustomization;

    case Kind.HelmRelease:
      return V2Routes.HelmRelease;

    case Kind.OCIRepository:
      return V2Routes.OCIRepository;

    case Kind.Provider:
      return V2Routes.Provider;

    default:
      break;
  }

  return '' as V2Routes;
}

type Props = {
  className?: string;
  onClick?: (o: DetailViewProps) => void;
  objects: FluxObject[];
  initialFilterState?: any;
};

function FluxObjectsTable({
  className,
  onClick,
  initialFilterState,
  objects,
}: Props) {
  const resolver = useLinkResolver();

  return (
    <DataTable
      filters={initialFilterState}
      className={className}
      fields={[
        {
          value: (u: FluxObject) => {
            // @ts-ignore
            const kind = Kind[u.type];
            const secret = u.type === 'Secret';
            const params = {
              name: u.name,
              namespace: u.namespace,
              clusterName: u.clusterName,
            };
            // Enterprise is "aware" of more types of objects than Core,
            // and we want to be able to link to those within this table.
            // The resolver func provided by the context will decide what URL this routes to.
            const resolved = resolver && resolver(u.type || '', params);
            const route = objectTypeToRoute(kind);
            const formatted = formatURL(route, params);

            if (route || resolved) {
              return <Link to={resolved || formatted}>{u.name}</Link>;
            }

            console.log(secret);

            return (
              <div
                onClick={() =>
                  secret
                    ? null
                    : onClick &&
                      onClick({
                        object: u,
                      })
                }
                color={secret ? 'neutral40' : 'primary10'}
                // pointer={!secret}
              >
                {u.name}
              </div>
            );
          },
          label: 'Name',
          sortValue: (u: FluxObject) => u.name || '',
          textSearchable: true,
          maxWidth: 600,
        },
        {
          label: 'Kind',
          value: (u: FluxObject) => u.type || '',
          sortValue: (u: FluxObject) => u.type,
        },
      ]}
      rows={objects}
    />
  );
}
export default styled(FluxObjectsTable).attrs({
  className: FluxObjectsTable.name,
})``;
