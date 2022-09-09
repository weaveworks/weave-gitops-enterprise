import {
  KustomizationDetail,
  useGetKustomization,
} from '@weaveworks/weave-gitops';
import { routeTab } from '@weaveworks/weave-gitops/ui/components/KustomizationDetail';
import { FC } from 'react';
import { useRouteMatch } from 'react-router-dom';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { FieldsType, PolicyViolationsList } from '../PolicyViolations/Table';
import { useApplicationsCount } from './utils';

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
  const applicationsCount = useApplicationsCount();
  const { data, isLoading, error } = useGetKustomization(
    name,
    namespace,
    clusterName,
  );
  const kustomization = data?.kustomization;
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
    <PageTemplate documentTitle="WeGO · Kustomization">
      <SectionHeader
        path={[
          {
            label: 'Applications',
            url: '/applications',
            count: applicationsCount,
          },
          {
            label: `${name}`,
          },
        ]}
      />
      <ContentWrapper
        loading={isLoading}
        errors={
          error ? [{ clusterName, namespace, message: error?.message }] : []
        }
      >
        <KustomizationDetail
          kustomization={kustomization}
          customTabs={customTabs}
        />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsKustomization;
