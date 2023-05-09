import {
  HelmReleaseDetail,
  Kind,
  useGetObject,
} from '@weaveworks/weave-gitops';
import { routeTab } from '@weaveworks/weave-gitops/ui/components/KustomizationDetail';
import { HelmRelease } from '@weaveworks/weave-gitops/ui/lib/objects';
import { FC } from 'react';
import { useRouteMatch } from 'react-router-dom';
import { Routes } from '../../utils/nav';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { FieldsType, PolicyViolationsList } from '../PolicyViolations/Table';
import { EditButton } from '../Templates/Edit/EditButton';

type Props = {
  name: string;
  clusterName: string;
  namespace: string;
};

const WGApplicationsHelmRelease: FC<Props> = props => {
  const { name, namespace, clusterName } = props;
  const {
    data: helmRelease,
    isLoading,
    error,
  } = useGetObject<HelmRelease>(name, namespace, Kind.HelmRelease, clusterName);

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
                kind: Kind.HelmRelease,
              }}
              tableType={FieldsType.application}
              sourcePath="helm_release"
            />
          </div>
        );
      },
      visible: true,
    },
  ];

  return (
    <PageTemplate
      documentTitle="Helm Release"
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
      <ContentWrapper
        loading={isLoading}
        errors={
          error ? [{ clusterName, namespace, message: error?.message }] : []
        }
      >
        {!error && !isLoading && (
          <HelmReleaseDetail
            helmRelease={helmRelease}
            customActions={[<EditButton resource={helmRelease} />]}
            customTabs={customTabs}
            {...props}
          />
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsHelmRelease;
