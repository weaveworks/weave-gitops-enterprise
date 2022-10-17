import {
  Kind,
  KustomizationDetail,
  useGetObject,
} from '@weaveworks/weave-gitops';
import { routeTab } from '@weaveworks/weave-gitops/ui/components/KustomizationDetail';
import { Kustomization } from '@weaveworks/weave-gitops/ui/lib/objects';
import { FC } from 'react';
import { useRouteMatch } from 'react-router-dom';
import { Routes } from '../../utils/nav';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { FieldsType, PolicyViolationsList } from '../PolicyViolations/Table';
import { EditButton } from '../EditButton';

type Props = {
  name: string;
  namespace: string;
  clusterName: string;
};

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
              req={{ clusterName, namespace, application: name }}
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
    <PageTemplate
      documentTitle="Kustomization"
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
        <KustomizationDetail
          kustomization={kustomization}
          customActions={[
            <EditButton resource={kustomization} isLoading={isLoading} />,
          ]}
          customTabs={customTabs}
        />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsKustomization;
