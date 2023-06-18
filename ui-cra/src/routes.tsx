import { V2Routes } from '@weaveworks/weave-gitops';
import qs from 'query-string';
import Lottie from 'react-lottie-player';
import { Redirect, Route, Switch } from 'react-router-dom';
import styled from 'styled-components';
import { GitProvider } from './api/gitauth/gitauth.pb';
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
import WGNotificationsProvider from './components/Applications/NotificationsProvider';
import WGApplicationsOCIRepository from './components/Applications/OCIRepository';
import WGApplicationsSources from './components/Applications/Sources';
import MCCP from './components/Clusters';
import ClusterDetails from './components/Clusters/ClusterDetails';
import Explorer from './components/Explorer';
import OAuthCallback from './components/GitAuth/OAuthCallback';
import GitOpsRunDetail from './components/GitOpsRun/Detail';
import GitOpsRun from './components/GitOpsRun/List';
import GitOpsSets from './components/GitOpsSets';
import GitOpsSetDetail from './components/GitOpsSets/GitOpsSetDetail';
import ImageAutomationPage from './components/ImageAutomation';
import ImagePolicyDetails from './components/ImageAutomation/policies/ImagePolicyDetails';
import ImageAutomationRepoDetails from './components/ImageAutomation/repositories/ImageAutomationRepoDetails';
import ImageAutomationUpdatesDetails from './components/ImageAutomation/updates/ImageAutomationUpdatesDetails';
import Pipelines from './components/Pipelines';
import PipelineDetails from './components/Pipelines/PipelineDetails';
import Policies from './components/Policies/PoliciesListPage';
import PolicyDetails from './components/Policies/PolicyDetailsPage';
import PolicyConfigsList from './components/PolicyConfigs';
import PolicyConfigsDetails from './components/PolicyConfigs/PolicyConfigDetails';
import CreatePolicyConfig from './components/PolicyConfigs/create';
import ProgressiveDelivery from './components/ProgressiveDelivery';
import CanaryDetails from './components/ProgressiveDelivery/CanaryDetails';
import SecretsList from './components/Secrets';
import CreateSecret from './components/Secrets/Create';
import CreateSOPS from './components/Secrets/SOPS';
import SecretDetails from './components/Secrets/SecretDetails';
import TemplatesDashboard from './components/Templates';
import AddClusterWithCredentials from './components/Templates/Create';
import EditResourcePage from './components/Templates/Edit';
import TerraformObjectDetail from './components/Terraform/TerraformObjectDetail';
import TerraformObjectList from './components/Terraform/TerraformObjectList';
import WGUserInfo from './components/UserInfo';
import Workspaces from './components/Workspaces';
import WorkspaceDetails from './components/Workspaces/WorkspaceDetails';
import { Routes } from './utils/nav';
import { NotificationsWrapper } from './components/Layout/NotificationsWrapper';
import PolicyViolationPage from './components/Policies/PolicyViolationPage';
import { Page } from './components/Layout/App';

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
  width: 100%;
`;

const Page404 = () => (
  <Page path={[{ label: 'Error' }]}>
    <NotificationsWrapper>
      <Lottie
        loop
        animationData={error404}
        play
        style={{ width: '100%', height: 650 }}
      />
    </NotificationsWrapper>
  </Page>
);

const AppRoutes = () => {
  return (
    <Switch>
      <Route exact path="/">
        <Redirect to={Routes.Clusters} />
      </Route>
      <Route component={MCCP} path={Routes.Clusters} />
      <Route component={MCCP} exact path={Routes.DeleteCluster} />
      <Route
        component={withSearchParams((props: any) => (
          <ClusterDetails {...props} />
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
        component={withSearchParams(PolicyViolationPage)}
        exact
        path={V2Routes.PolicyViolationDetails}
      />
      <Route component={GitOpsRun} exact path={Routes.GitOpsRun} />
      <Route
        component={withSearchParams(GitOpsRunDetail)}
        path={Routes.GitOpsRunDetail}
      />
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
        component={(props: any) => (
          <CoreWrapper>
            <WGApplicationsSources {...props} />
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
            <WGUserInfo {...props} />
          </CoreWrapper>
        ))}
        path={V2Routes.UserInfo}
      />
      <Route
        component={withSearchParams(PolicyViolationPage)}
        exact
        path={V2Routes.PolicyViolationDetails}
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
      <Route path={Routes.ImageAutomation} component={ImageAutomationPage} />
      <Route
        path={V2Routes.ImageAutomationUpdatesDetails}
        component={withSearchParams(ImageAutomationUpdatesDetails)}
      />
      <Route
        path={V2Routes.ImageAutomationRepositoryDetails}
        component={withSearchParams(ImageAutomationRepoDetails)}
      />
      <Route
        path={V2Routes.ImagePolicyDetails}
        component={withSearchParams(ImagePolicyDetails)}
      />
      <Route exact path={V2Routes.Policies} component={Policies} />
      <Route
        component={withSearchParams(PolicyDetails)}
        path={V2Routes.PolicyDetailsPage}
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
      <Route exact path={Routes.CreateSecret} component={CreateSecret} />
      <Route exact path={Routes.CreateSopsSecret} component={CreateSOPS} />
      <Route exact path={Routes.PolicyConfigs} component={PolicyConfigsList} />
      <Route
        exact
        path={Routes.PolicyConfigsDetails}
        component={withSearchParams(PolicyConfigsDetails)}
      />
      <Route
        exact
        path={Routes.CreatePolicyConfig}
        component={CreatePolicyConfig}
      />

      <Route
        path={Routes.TerraformDetail}
        component={withSearchParams(TerraformObjectDetail)}
      />
      <Route path={Routes.Explorer} component={withSearchParams(Explorer)} />
      <Route
        exact
        path={Routes.GitOpsSets}
        component={withSearchParams(GitOpsSets)}
      />
      <Route
        path={Routes.GitOpsSetDetail}
        component={withSearchParams(GitOpsSetDetail)}
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
              state=""
            />
          );
        }}
      />
      <Route
        exact
        path={Routes.BitBucketOauthCallback}
        component={({ location }: any) => {
          const params = qs.parse(location.search);

          const error = Array.isArray(params?.error)
            ? params?.error.join(', ')
            : params?.error;

          const desc = Array.isArray(params.error_description)
            ? params.error_description?.join('\n')
            : params?.error_description;

          return (
            <OAuthCallback
              provider={GitProvider.BitBucketServer}
              code={params.code as string}
              state={params.state as string}
              error={error}
              errorDescription={desc}
            />
          );
        }}
      />
      <Route
        exact
        path={Routes.AzureDevOpsOauthCallback}
        component={({ location }: any) => {
          const params = qs.parse(location.search);

          const error = Array.isArray(params?.error)
            ? params?.error.join(', ')
            : params?.error;

          const desc = Array.isArray(params.error_description)
            ? params.error_description?.join('\n')
            : params?.error_description;

          return (
            <OAuthCallback
              provider={GitProvider.AzureDevOps}
              code={params.code as string}
              state={params.state as string}
              error={error}
              errorDescription={desc}
            />
          );
        }}
      />
      <Route exact render={Page404} />
    </Switch>
  );
};

export default AppRoutes;
