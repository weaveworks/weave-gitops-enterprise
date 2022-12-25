import { V2Routes } from '@weaveworks/weave-gitops';
import qs from 'query-string';
import Lottie from 'react-lottie-player';
import { Redirect, Route, Switch } from 'react-router-dom';
import styled from 'styled-components';
import error404 from './assets/img/error404.json';
import WGApplicationsDashboard from './components/Applications';
import AddApplication from './components/Applications/Add';
import WGApplicationsBucket from './components/Applications/Bucket';
import WGApplicationsFluxRuntime from './components/Applications/FluxRuntime';
import WGApplicationsGitRepository from './components/Applications/GitRepository';
import WGApplicationsHelmChart from './components/Applications/HelmChart';
import WGApplicationsHelmRelease from './components/Applications/HelmRelease';
import WGApplicationsHelmRepository from './components/Applications/HelmRepository';
import WGApplicationsKustomization from './components/Applications/Kustomization';
import WGNotifications from './components/Applications/Notifications';
import WGApplicationsOCIRepository from './components/Applications/OCIRepository';
import WGNotificationsProvider from './components/Applications/NotificationsProvider';
import WGApplicationsSources from './components/Applications/Sources';
import MCCP from './components/Clusters';
import ClusterDashboard from './components/Clusters/ClusterDashboard';
import GitOpsRun from './components/GitOpsRun';
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
import AddClusterWithCredentials from './components/Templates/Create';
import EditResourcePage from './components/Templates/Edit';
import TerraformObjectDetail from './components/Terraform/TerraformObjectDetail';
import TerraformObjectList from './components/Terraform/TerraformObjectList';
import Workspaces from './components/Workspaces';
import WorkspaceDetails from './components/Workspaces/WorkspaceDetails';
import { Routes } from './utils/nav';
import OAuthCallback from './components/GithubAuth/OAuthCallback';
import { GitProvider } from './api/gitauth/gitauth.pb';
import SecretsList from './components/Secrets';
import SecretDetails from './components/Secrets/SecretDetails';

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
        <Redirect to={Routes.Clusters} />
      </Route>
      <Route component={MCCP} exact path={Routes.Clusters} />
      <Route component={MCCP} exact path={Routes.DeleteCluster} />
      <Route
        component={withSearchParams((props: any) => (
          <ClusterDashboard {...props} />
        ))}
        path={Routes.ClusterDashboard}
      />
      <Route
        component={withSearchParams((props: any) => (
          <CoreWrapper>
            <EditResourcePage {...props} />
          </CoreWrapper>
        ))}
        path={Routes.EditResource}
      />
      <Route
        component={AddClusterWithCredentials}
        exact
        path={Routes.AddCluster}
      />
      <Route
        component={PoliciesViolations}
        exact
        path={Routes.PolicyViolations}
      />
      <Route
        component={withSearchParams(PolicyViolationDetails)}
        exact
        path={Routes.PolicyViolationDetails}
      />
      <Route component={GitOpsRun} exact path={Routes.GitOpsRun} />
      <Route
        component={(props: any) => (
          <CoreWrapper>
            <WGApplicationsDashboard {...props} />
          </CoreWrapper>
        )}
        exact
        path={V2Routes.Automations}
      />
      <Route
        component={withSearchParams(AddApplication)}
        exact
        path={Routes.AddApplication}
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
      <Route
        component={withSearchParams((props: any) => (
          <CoreWrapper>
            <WGNotificationsProvider {...props} />
          </CoreWrapper>
        ))}
        path={V2Routes.Provider}
      />
      <Route exact path={Routes.Canaries} component={ProgressiveDelivery} />
      <Route
        path={Routes.CanaryDetails}
        component={withSearchParams(CanaryDetails)}
      />
      <Route exact path={Routes.Pipelines} component={Pipelines} />
      <Route
        path={Routes.PipelineDetails}
        component={withSearchParams(PipelineDetails)}
      />

      <Route exact path={Routes.Policies} component={Policies} />
      <Route
        exact
        path={Routes.PolicyDetails}
        component={withSearchParams(PolicyDetails)}
      />
      <Route component={TemplatesDashboard} exact path={Routes.Templates} />
      <Route
        exact
        path={Routes.TerraformObjects}
        component={withSearchParams(TerraformObjectList)}
      />
      <Route exact path={Routes.Workspaces} component={Workspaces} />
      <Route
        path={Routes.WorkspaceDetails}
        component={withSearchParams(WorkspaceDetails)}
      />

      <Route exact path={Routes.Secrets} component={SecretsList} />
      <Route
        path={Routes.SecretDetails}
        component={withSearchParams(SecretDetails)}
      />
      <Route
        path={Routes.TerraformDetail}
        component={withSearchParams(TerraformObjectDetail)}
      />
      <Route
        exact
        path={Routes.GitlabOauthCallback}
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
