import { V2Routes, OAuthCallback } from '@weaveworks/weave-gitops';
import { GitProvider } from '@weaveworks/weave-gitops/ui/lib/api/applications/applications.pb';
import qs from 'qs';
import Lottie from 'react-lottie-player';
import { Switch, Route, Redirect } from 'react-router-dom';
import styled from 'styled-components';
import WGApplicationsDashboard from './components/Applications';
import WGApplicationsBucket from './components/Applications/Bucket';
import WGApplicationsFluxRuntime from './components/Applications/FluxRuntime';
import WGApplicationsGitRepository from './components/Applications/GitRepository';
import WGApplicationsHelmChart from './components/Applications/HelmChart';
import WGApplicationsHelmRelease from './components/Applications/HelmRelease';
import WGApplicationsHelmRepository from './components/Applications/HelmRepository';
import WGApplicationsKustomization from './components/Applications/Kustomization';
import WGNotifications from './components/Applications/Notifications';
import WGApplicationsOCIRepository from './components/Applications/OCIRepository';
import WGApplicationsSources from './components/Applications/Sources';
import MCCP from './components/Clusters';
import ClusterDashboard from './components/Clusters/ClusterDashboard';
import EditClusterPage from './components/Clusters/Edit';
import { ContentWrapper } from './components/Layout/ContentWrapper';
import { PageTemplate } from './components/Layout/PageTemplate';
import Pipelines from './components/Pipelines';
import PipelineDetails from './components/Pipelines/PipelineDetails';
import Policies from './components/Policies';
import PolicyDetails from './components/Policies/PolicyDetails';
import PoliciesViolations from './components/PolicyViolations';
import PolicyViolationDetails from './components/PolicyViolations/ViolationDetails';
import ProgressiveDelivery from './components/ProgressiveDelivery';
import CanaryDetails from './components/ProgressiveDelivery/CanaryDetails';
import TemplatesDashboard from './components/Templates';
import TerraformObjectDetail from './components/Terraform/TerraformObjectDetail';
import TerraformObjectList from './components/Terraform/TerraformObjectList';
import { Routes } from './utils/nav';
import AddApplication from './components/Applications/Add';
import AddClusterWithCredentials from './components/Clusters/Create';
import error404 from './assets/img/error404.json';

const ROUTES = {
  CLUSTERS: '/clusters',
  CLUSTER: '/cluster',
  DELETE_CLUSTER: '/clusters/delete',
  GITLAB_OAUTH_CALLBACK: '/oauth/gitlab',
  POLICIES: '/policies',
  CANARIES: '/applications/delivery',
  PIPELINES: '/applications/pipelines',
  EDIT_CLUSTER: '/clusters/:clusterName/edit',
};

function withSearchParams(Cmp: any) {
  return ({ location: { search }, ...rest }: any) => {
    const params = qs.parse(search);

    return <Cmp {...rest} {...params} />;
  };
}

const CoreWrapper = styled.div`
  div[class*='FilterDialog__SlideContainer'] {
    overflow: hidden;
  }
  .MuiFormControl-root {
    min-width: 0px;
  }
  div[class*='ReconciliationGraph'] {
    svg {
      min-height: 600px;
    }
    .MuiSlider-root.MuiSlider-vertical {
      height: 200px;
    }
  }
  .MuiButton-root {
    margin-right: 0;
  }
  max-width: calc(100vw - 220px);
`;

const Page404 = () => (
  <PageTemplate documentTitle="NotFound" path={[{ label: 'Error' }]}>
    <ContentWrapper>
      <Lottie
        loop
        animationData={error404}
        play
        style={{ width: '100%', height: 650 }}
      />
    </ContentWrapper>
  </PageTemplate>
);

const AppRoutes = () => {
  return (
    <Switch>
      <Route exact path="/">
        <Redirect to={ROUTES.CLUSTERS} />
      </Route>
      <Route component={MCCP} exact path={ROUTES.CLUSTERS} />
      <Route component={MCCP} exact path={ROUTES.DELETE_CLUSTER} />
      <Route
        component={withSearchParams((props: any) => (
          <ClusterDashboard {...props} />
        ))}
        path={ROUTES.CLUSTER}
      />
      <Route component={EditClusterPage} exact path={ROUTES.EDIT_CLUSTER} />
      <Route
        component={AddClusterWithCredentials}
        exact
        path="/templates/:templateName/create"
      />
      <Route component={PoliciesViolations} exact path="/clusters/violations" />
      <Route
        component={withSearchParams(PolicyViolationDetails)}
        exact
        path="/clusters/violations/details"
      />
      <Route
        component={() => (
          <CoreWrapper>
            <WGApplicationsDashboard />
          </CoreWrapper>
        )}
        exact
        path={V2Routes.Automations}
      />
      <Route
        component={withSearchParams(AddApplication)}
        exact
        path="/applications/create"
      />
      <Route
        component={() => (
          <CoreWrapper>
            <WGApplicationsSources />
          </CoreWrapper>
        )}
        exact
        path={V2Routes.Sources}
      />
      <Route
        component={withSearchParams((props: any) => (
          <CoreWrapper>
            <WGApplicationsKustomization {...props} />
          </CoreWrapper>
        ))}
        path={V2Routes.Kustomization}
      />
      <Route
        component={withSearchParams((props: any) => (
          <CoreWrapper>
            <WGApplicationsGitRepository {...props} />
          </CoreWrapper>
        ))}
        path={V2Routes.GitRepo}
      />
      <Route
        component={withSearchParams((props: any) => (
          <CoreWrapper>
            <WGApplicationsHelmRepository {...props} />
          </CoreWrapper>
        ))}
        path={V2Routes.HelmRepo}
      />
      <Route
        component={withSearchParams((props: any) => (
          <CoreWrapper>
            <WGApplicationsBucket {...props} />
          </CoreWrapper>
        ))}
        path={V2Routes.Bucket}
      />
      <Route
        component={withSearchParams((props: any) => (
          <CoreWrapper>
            <WGApplicationsHelmRelease {...props} />
          </CoreWrapper>
        ))}
        path={V2Routes.HelmRelease}
      />
      <Route
        component={withSearchParams((props: any) => (
          <CoreWrapper>
            <WGApplicationsHelmChart {...props} />
          </CoreWrapper>
        ))}
        path={V2Routes.HelmChart}
      />
      <Route
        component={withSearchParams((props: any) => (
          <CoreWrapper>
            <WGApplicationsOCIRepository {...props} />
          </CoreWrapper>
        ))}
        path={V2Routes.OCIRepository}
      />
      <Route
        component={() => (
          <CoreWrapper>
            <WGApplicationsFluxRuntime />
          </CoreWrapper>
        )}
        path={V2Routes.FluxRuntime}
      />
      <Route
        component={withSearchParams((props: any) => (
          <CoreWrapper>
            <WGNotifications {...props} />
          </CoreWrapper>
        ))}
        path={V2Routes.Notifications}
      />
      <Route exact path={ROUTES.CANARIES} component={ProgressiveDelivery} />
      <Route exact path={ROUTES.PIPELINES} component={Pipelines} />
      <Route
        exact
        path="/applications/pipelines/details"
        component={withSearchParams(PipelineDetails)}
      />
      <Route
        path="/applications/delivery/:id"
        component={withSearchParams(CanaryDetails)}
      />
      <Route exact path={ROUTES.POLICIES} component={Policies} />
      <Route
        exact
        path="/policies/details"
        component={withSearchParams(PolicyDetails)}
      />
      <Route component={TemplatesDashboard} exact path="/templates" />
      <Route
        exact
        path={Routes.TerraformObjects}
        component={withSearchParams(TerraformObjectList)}
      />
      <Route
        path={Routes.TerraformDetail}
        component={withSearchParams(TerraformObjectDetail)}
      />
      <Route
        exact
        path={ROUTES.GITLAB_OAUTH_CALLBACK}
        component={({ location }: any) => {
          const params = qs.parse(location.search);
          return (
            <OAuthCallback
              provider={'GitLab' as GitProvider}
              code={params.code as string}
            />
          );
        }}
      />
      <Route exact render={Page404} />
    </Switch>
  );
};

export default AppRoutes;
