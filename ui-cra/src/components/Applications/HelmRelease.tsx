import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { useApplicationsCount } from './utils';
import {
  HelmReleaseDetail,
  LoadingPage,
  useGetHelmRelease,
} from '@weaveworks/weave-gitops';
import { routeTab } from '@weaveworks/weave-gitops/ui/components/KustomizationDetail';
import { useRouteMatch } from 'react-router-dom';
import { FieldsType, PolicyViolationsList } from '../PolicyViolations/Table';

type Props = {
  name: string;
  clusterName: string;
  namespace: string;
};

const WGApplicationsHelmRelease: FC<Props> = props => {
  const applicationsCount = useApplicationsCount();
  const { name, namespace, clusterName } = props;
  const { data, isLoading, error } = useGetHelmRelease(
    name,
    namespace,
    clusterName,
  );
  const helmRelease = data?.helmRelease;

  const { path } = useRouteMatch();
  const customTabs: Array<routeTab> = [
    {
      name: 'Violations',
      path: `${path}/violations`,
      component: () => {
        return (
          <PolicyViolationsList
            req={{ clusterName, namespace }}
            tableType={FieldsType.application}
          />
        );
      },
      visible: true,
    },
  ];
  return (
    <PageTemplate documentTitle="WeGO · Helm Release">
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
        {error && <h3>{error.message}</h3>}
        {isLoading && <LoadingPage />}
        {!error && !isLoading && (
          <HelmReleaseDetail helmRelease={helmRelease} {...props} customTabs={customTabs}/>
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsHelmRelease;
