import { V2Routes } from '@weaveworks/weave-gitops';
import qs from 'query-string';
import Lottie from 'react-lottie-player';
import {
  Navigate,
  Route,
  Routes as Routess,
  useLocation,
} from 'react-router-dom';
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
import { ContentWrapper } from './components/Layout/ContentWrapper';
import { PageTemplate } from './components/Layout/PageTemplate';
import Pipelines from './components/Pipelines';
import PipelineDetails from './components/Pipelines/PipelineDetails';
import Policies from './components/Policies';
import PolicyDetails from './components/Policies/PolicyDetails';
import PolicyConfigsList from './components/PolicyConfigs';
import PolicyConfigsDetails from './components/PolicyConfigs/PolicyConfigDetails';
import CreatePolicyConfig from './components/PolicyConfigs/create';
// import PoliciesViolations from './components/PolicyViolations';
import PolicyViolationDetails from './components/PolicyViolations/ViolationDetails';
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

function WithSearchParams() {
  const location = useLocation();
  const params = qs.parse(location.search);
  // FIXME one day
  return params as any;
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
    <Routess>
      <Route
        path="/"
        element={<Navigate to={Routes.Clusters + '/*'} replace />}
      />
      <Route element={<MCCP />} path={Routes.Clusters + '/*'} />
      <Route element={<MCCP />} path={Routes.DeleteCluster} />
      <Route
        element={<ClusterDetails {...WithSearchParams()} />}
        path={Routes.ClusterDashboard + '/*'}
      />
      <Route element={<AddClusterWithCredentials />} path={Routes.AddCluster} />
      <Route element={<TemplatesDashboard />} path={Routes.Templates} />
      <Route
        path={Routes.GitOpsSets}
        element={<GitOpsSets {...WithSearchParams()} />}
      />
      <Route
        path={Routes.GitOpsSetDetail + '/*'}
        element={<GitOpsSetDetail {...WithSearchParams()} />}
      />
      <Route
        path={Routes.TerraformObjects}
        element={<TerraformObjectList {...WithSearchParams()} />}
      />
      <Route
        path={Routes.TerraformDetail + '/*'}
        element={<TerraformObjectDetail {...WithSearchParams()} />}
      />
      <Route path={Routes.Secrets} element={<SecretsList />} />
      <Route
        path={Routes.SecretDetails + '/*'}
        element={<SecretDetails {...WithSearchParams()} />}
      />
      <Route path={Routes.CreateSecret} element={<CreateSecret />} />
      <Route path={Routes.CreateSopsSecret} element={<CreateSOPS />} />

      <Route
        element={
          <CoreWrapper>
            <WGApplicationsDashboard {...WithSearchParams()} />
          </CoreWrapper>
        }
        path={V2Routes.Automations}
      />
      <Route
        element={<AddApplication {...WithSearchParams()} />}
        path={Routes.AddApplication}
      />
      <Route
        element={
          <CoreWrapper>
            <WGApplicationsKustomization {...WithSearchParams()} />
          </CoreWrapper>
        }
        path={V2Routes.Kustomization + '/*'}
      />
      <Route
        element={
          <CoreWrapper>
            <WGApplicationsHelmRelease {...WithSearchParams()} />
          </CoreWrapper>
        }
        path={V2Routes.HelmRelease + '/*'}
      />
      <Route
        element={
          <CoreWrapper>
            <WGApplicationsSources {...WithSearchParams()} />
          </CoreWrapper>
        }
        path={V2Routes.Sources}
      />
      <Route
        element={
          <CoreWrapper>
            <WGApplicationsHelmRepository {...WithSearchParams()} />
          </CoreWrapper>
        }
        path={V2Routes.HelmRepo + '/*'}
      />
      <Route
        element={
          <CoreWrapper>
            <WGApplicationsHelmChart {...WithSearchParams()} />
          </CoreWrapper>
        }
        path={V2Routes.HelmChart + '/*'}
      />
      <Route
        element={
          <CoreWrapper>
            <WGApplicationsBucket {...WithSearchParams()} />
          </CoreWrapper>
        }
        path={V2Routes.Bucket + '/*'}
      />
      <Route
        element={
          <CoreWrapper>
            <WGApplicationsGitRepository {...WithSearchParams()} />
          </CoreWrapper>
        }
        path={V2Routes.GitRepo + '/*'}
      />
      <Route
        element={
          <CoreWrapper>
            <WGApplicationsOCIRepository {...WithSearchParams()} />
          </CoreWrapper>
        }
        path={V2Routes.OCIRepository + '/*'}
      />
      <Route
        element={
          <CoreWrapper>
            <WGApplicationsFluxRuntime />
          </CoreWrapper>
        }
        path={V2Routes.FluxRuntime + '/*'}
      />
      <Route path={Routes.ImageAutomation} element={<ImageAutomationPage />} />
      <Route
        path={V2Routes.ImageAutomationUpdatesDetails + '/*'}
        element={<ImageAutomationUpdatesDetails {...WithSearchParams()} />}
      />
      <Route
        path={V2Routes.ImageAutomationRepositoryDetails + '/*'}
        element={<ImageAutomationRepoDetails {...WithSearchParams()} />}
      />
      <Route
        path={V2Routes.ImagePolicyDetails + '/*'}
        element={<ImagePolicyDetails {...WithSearchParams()} />}
      />
      <Route path={Routes.Pipelines} element={<Pipelines />} />
      <Route
        path={Routes.PipelineDetails + '/*'}
        element={<PipelineDetails {...WithSearchParams()} />}
      />
      <Route path={Routes.Canaries} element={<ProgressiveDelivery />} />
      <Route
        path={Routes.CanaryDetails + '/*'}
        element={<CanaryDetails {...WithSearchParams()} />}
      />
      <Route path={Routes.Workspaces} element={<Workspaces />} />
      <Route
        path={Routes.WorkspaceDetails + '/*'}
        element={<WorkspaceDetails {...WithSearchParams()} />}
      />
      <Route
        path={Routes.Explorer + '/*'}
        element={<Explorer {...WithSearchParams()} />}
      />
      <Route path={Routes.Policies} element={<Policies />} />
      <Route
        path={Routes.PolicyDetails}
        element={<PolicyDetails {...WithSearchParams()} />}
      />
      <Route path={Routes.PolicyConfigs} element={<PolicyConfigsList />} />
      <Route
        path={Routes.PolicyConfigsDetails}
        element={<PolicyConfigsDetails {...WithSearchParams()} />}
      />
      <Route
        path={Routes.CreatePolicyConfig}
        element={<CreatePolicyConfig />}
      />
      {/* <Route element={<PoliciesViolations />} path={Routes.PolicyViolations} /> */}
      <Route
        element={<PolicyViolationDetails {...WithSearchParams()} />}
        path={Routes.PolicyViolationDetails}
      />
      <Route element={<GitOpsRun />} path={Routes.GitOpsRun} />
      <Route
        element={<GitOpsRunDetail {...WithSearchParams()} />}
        path={Routes.GitOpsRunDetail + '/*'}
      />
      <Route
        element={
          <CoreWrapper>
            <WGNotifications {...WithSearchParams()} />
          </CoreWrapper>
        }
        path={V2Routes.Notifications}
      />
      <Route
        element={
          <CoreWrapper>
            <WGNotificationsProvider {...WithSearchParams()} />
          </CoreWrapper>
        }
        path={V2Routes.Provider}
      />
      <Route
        element={
          <CoreWrapper>
            <EditResourcePage {...WithSearchParams()} />
          </CoreWrapper>
        }
        path={Routes.EditResource}
      />
      <Route
        element={
          <CoreWrapper>
            <WGUserInfo {...WithSearchParams()} />
          </CoreWrapper>
        }
        path={V2Routes.UserInfo}
      />
      <Route
        path={Routes.GitlabOauthCallback}
        element={
          <OAuthCallback
            provider={'GitLab' as GitProvider}
            code={WithSearchParams().code as string}
            state=""
          />
        }
      />

      <Route path="*" element={<Page404 />} />
    </Routess>
  );
};

export default AppRoutes;
