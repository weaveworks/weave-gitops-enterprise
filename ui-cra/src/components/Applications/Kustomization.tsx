import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { useApplicationsCount } from './utils';
import {
  KustomizationDetail,
  useGetKustomization,
} from '@weaveworks/weave-gitops';
import { routeTab } from '@weaveworks/weave-gitops/ui/components/KustomizationDetail';
import { useRouteMatch } from 'react-router-dom';
import { FieldsType, PolicyViolationsList } from '../PolicyViolations/Table';

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
  const { data } = useGetKustomization(name, namespace, clusterName);
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
              req={{ clusterName, namespace }}
              tableType={FieldsType.application}
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
      <ContentWrapper>
        <KustomizationDetail
          kustomization={kustomization}
          customTabs={customTabs}
        />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsKustomization;
