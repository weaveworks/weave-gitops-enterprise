import { HelmReleaseDetail, useGetHelmRelease } from '@weaveworks/weave-gitops';
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
          <div style={{ width: '100%' }}>
            <PolicyViolationsList
              req={{ clusterName, namespace, application: name }}
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
      <ContentWrapper
        loading={isLoading}
        errors={[{ clusterName, namespace, message: error?.message }]}
      >
        {!error && !isLoading && (
          <HelmReleaseDetail
            helmRelease={helmRelease}
            {...props}
            customTabs={customTabs}
          />
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsHelmRelease;
