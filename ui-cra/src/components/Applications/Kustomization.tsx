import {
  formatURL,
  Kind,
  KustomizationDetail,
  LinkResolverProvider,
  useGetObject,
} from '@weaveworks/weave-gitops';
import { routeTab } from '@weaveworks/weave-gitops/ui/components/KustomizationDetail';
import { Kustomization } from '@weaveworks/weave-gitops/ui/lib/objects';
import { FC } from 'react';
import { useRouteMatch } from 'react-router-dom';
import { Routes } from '../../utils/nav';
import { formatClusterDashboardUrl } from '../Clusters/ClusterDashboardLink';
import { FieldsType, PolicyViolationsList } from '../PolicyViolations/Table';
import { EditButton } from '../Templates/Edit/EditButton';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import { Page } from '../Layout/App';

type Props = {
  name: string;
  namespace: string;
  clusterName: string;
};

function resolveLink(obj: string, params: any) {
  switch (obj) {
    case 'Canary':
      return formatURL(Routes.CanaryDetails, params);

    case 'Pipeline':
      return formatURL(Routes.PipelineDetails, params);
    case 'ClusterDashboard':
      return formatClusterDashboardUrl(params.clusterName);
    case 'Terraform':
      return formatURL(Routes.TerraformDetail, params);
    default:
      return null;
  }
}

const WGApplicationsKustomization: FC<Props> = ({
  name,
  namespace,
  clusterName,
}) => {
  const {
    data: kustomization,
    isLoading,
    error,
  } = useGetObject<Kustomization>(
    name,
    namespace,
    Kind.Kustomization,
    clusterName,
  );
  const { path } = useRouteMatch();

  const customTabs: Array<routeTab> = [
    {
      name: 'Violations',
      path: `${path}/violations`,
      component: () => {
        return (
          <div style={{ width: '100%' }}>
            <PolicyViolationsList
              req={{
                clusterName,
                namespace,
                application: name,
                kind: Kind.Kustomization,
              }}
              tableType={FieldsType.application}
              sourcePath="kustomization"
            />
          </div>
        );
      },
      visible: true,
    },
  ];

  return (
    <Page
      loading={isLoading}
      path={[
        {
          label: 'Applications',
          url: Routes.Applications,
        },
        {
          label: `${name}`,
        },
      ]}
    >
      <NotificationsWrapper
        errors={
          error ? [{ clusterName, namespace, message: error?.message }] : []
        }
      >
        <LinkResolverProvider
          resolver={(obj, params) => {
            const resolved = resolveLink(obj, {
              clusterName: params.clusterName,
              namespace: params.namespace,
              name: params.name,
            });
            return resolved || '';
          }}
        >
          <KustomizationDetail
            kustomization={kustomization}
            customActions={[<EditButton resource={kustomization} />]}
            customTabs={customTabs}
          />
        </LinkResolverProvider>
      </NotificationsWrapper>
    </Page>
  );
};

export default WGApplicationsKustomization;
