import React, { useCallback, useEffect, useState } from 'react';
import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { AddApplicationRequest, useApplicationsCount } from '../utils';
import useClusters from '../../../contexts/Clusters';
import GitOps from '../../Clusters/Create/Form/Partials/GitOps';
import { Grid, MenuItem } from '@material-ui/core';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import useTemplates from '../../../contexts/Templates';
import {
  CallbackStateContextProvider,
  clearCallbackState,
  getCallbackState,
  getProviderToken,
} from '@weaveworks/weave-gitops';
import { useHistory } from 'react-router-dom';
import { theme as weaveTheme } from '@weaveworks/weave-gitops';
import { isUnauthenticated, removeToken } from '../../../utils/request';
import useNotifications from '../../../contexts/Notifications';
import { GitProvider } from '@weaveworks/weave-gitops/ui/lib/api/applications/applications.pb';
import { useListConfig } from '../../../hooks/versions';
import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';
import styled from 'styled-components';
import { Input, Select } from '../../../utils/form';
import { useListGitRepos } from '../../../hooks/gitReposSource';
import _ from 'lodash';

const FormWrapper = styled.form`
  .form-section {
    width: 50%;
  }
`;

const AddApplication = () => {
  const applicationsCount = useApplicationsCount();
  const { clusters, isLoading } = useClusters();
  const [loading, setLoading] = useState<boolean>(false);
  const [showAuthDialog, setShowAuthDialog] = useState(false);
  const history = useHistory();
  const { setNotifications } = useNotifications();
  const { data } = useListConfig();
  const repositoryURL = data?.repositoryURL || '';
  const authRedirectPage = `/applications/new`;
  const { data: GitRepoResponse } = useListGitRepos();

  let initialFormData = {
    url: '',
    provider: '',
    branchName: `add-application-branch`,
    title: 'Add application',
    commitMessage: 'add application',
    pullRequestDescription: 'This PR add a new application',
    clusterKustomizations: [{}],
    name: '',
    namespace: '',
    cluster_name: '',
    cluster_namespace: '',
    cluster: '',
    cluster_isControlPlane: false,
    path: '',
    source_name: '',
    source_namespace: '',
    source: '',
  };

  const callbackState = getCallbackState();

  if (callbackState) {
    initialFormData = {
      ...initialFormData,
      ...callbackState.state.formData,
    };
  }
  const [formData, setFormData] = useState<any>(initialFormData);

  useEffect(() => {
    if (repositoryURL != null) {
      setFormData((prevState: any) => ({
        ...prevState,
        url: repositoryURL,
      }));
    }
  }, [repositoryURL]);

  useEffect(() => {
    clearCallbackState();
  }, []);

  useEffect(() => {
    setFormData((prevState: any) => ({
      ...prevState,
      pullRequestTitle: `Add application ${formData.name || ''}`,
    }));
  }, [formData.name]);

  const handleAddApplication = useCallback(() => {
    const payload = {
      head_branch: formData.branchName,
      title: formData.pullRequestTitle,
      description: formData.pullRequestDescription,
      commit_message: formData.commitMessage,
      clusterKustomizations: [
        {
          cluster: {
            name: formData.cluster_name,
            namespace: formData.cluster_namespace,
          },
          isControlPlane: formData.cluster_isControlPlane,
          kustomization: {
            metadata: {
              name: formData.name,
              namespace: formData.namespace,
            },
            spec: {
              path: formData.path,
              sourceRef: {
                name: formData.source_name,
                namespace: formData.source_namespace,
              },
            },
          },
        },
      ],
    };
    setLoading(true);
    return AddApplicationRequest(
      payload,
      getProviderToken(formData.provider as GitProvider),
    )
      .then(response => {
        history.push('/applications');
        setNotifications([
          {
            message: {
              component: (
                <a
                  style={{ color: weaveTheme.colors.primary }}
                  href={response.webUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  PR created successfully.
                </a>
              ),
            },
            variant: 'success',
          },
        ]);
      })
      .catch(error => {
        setNotifications([
          { message: { text: error.message }, variant: 'danger' },
        ]);
        if (isUnauthenticated(error.code)) {
          removeToken(formData.provider);
        }
      })
      .finally(() => setLoading(false));
  }, [formData, history, setNotifications]);

  const handleSelectCluster = (event: React.ChangeEvent<any>) => {
    const value = event.target.value;
    setFormData({
      ...formData,
      cluster_name: JSON.parse(value).name,
      cluster_namespace: JSON.parse(value).namespace,
      cluster_isControlPlane: JSON.parse(value).controlPlane,
      cluster: value,
    });
  };
  const clusterName = formData.cluster_namespace
    ? `${formData.cluster_namespace}/${formData.cluster_name}`
    : `${formData.cluster_name}`;
  const gitResposFilterdList = _.filter(GitRepoResponse?.gitRepositories, [
    'clusterName',
    clusterName,
  ]);

  const handleSelectSource = (event: React.ChangeEvent<any>) => {
    const value = event.target.value;
    setFormData({
      ...formData,
      source_name: JSON.parse(value).name,
      source_namespace: JSON.parse(value).namespace,
      source: value,
    });
  };

  const handleFormData = (
    event:
      | React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>
      | React.ChangeEvent<{ name?: string; value: unknown }>,
    fieldName?: string,
  ) => {
    const { value } = event?.target;
    setFormData({ ...formData, [fieldName as string]: value });
  };

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Add new application">
        <CallbackStateContextProvider
          callbackState={{
            page: authRedirectPage as PageRoute,
            state: {
              formData,
            },
          }}
        >
          <SectionHeader
            className="count-header"
            path={[
              {
                label: 'Applications',
                url: '/applications',
                count: applicationsCount,
              },
              { label: 'Add new application' },
            ]}
          />
          <ContentWrapper>
            <Grid container>
              <Grid item xs={12} sm={10} md={10} lg={8}>
                <FormWrapper>
                  <Input
                    className="form-section"
                    required={true}
                    name="name"
                    label="APPLICATION NAME"
                    value={formData.name}
                    onChange={event => handleFormData(event, 'name')}
                    description="define application name"
                  />
                  <Input
                    className="form-section"
                    required={true}
                    name="namespace"
                    label="APPLICATION NAMESPACE"
                    value={formData.namespace}
                    onChange={event => handleFormData(event, 'namespace')}
                    description="define application namespace"
                  />
                  <Select
                    className="form-section"
                    name="cluster_name"
                    required={true}
                    label="SELECT CLUSTER"
                    value={formData.cluster || ''}
                    onChange={handleSelectCluster}
                    defaultValue={''}
                    description="select target cluster"
                  >
                    {clusters?.map((option: any) => {
                      return (
                        <MenuItem
                          key={option.name}
                          value={JSON.stringify(option)}
                        >
                          {option.name}
                        </MenuItem>
                      );
                    })}
                  </Select>
                  <Select
                    className="form-section"
                    name="source"
                    required={true}
                    label="SELECT SOURCE"
                    value={formData.source || ''}
                    onChange={handleSelectSource}
                    defaultValue={''}
                    description="The name and type of source"
                  >
                    {gitResposFilterdList?.map((option: any) => {
                      return (
                        <MenuItem
                          key={option.cluseterName}
                          value={JSON.stringify(option)}
                        >
                          {option.name}
                        </MenuItem>
                      );
                    })}
                  </Select>
                  <Input
                    className="form-section"
                    required={true}
                    name="path"
                    label="SELECT PATH/CHART"
                    value={formData.path}
                    onChange={event => handleFormData(event, 'path')}
                    description="The name of the path"
                  />
                </FormWrapper>
              </Grid>
              <Grid item xs={12} sm={10} md={10} lg={8}>
                <GitOps
                  loading={loading}
                  formData={formData}
                  setFormData={setFormData}
                  onSubmit={handleAddApplication}
                  showAuthDialog={showAuthDialog}
                  setShowAuthDialog={setShowAuthDialog}
                />
              </Grid>
            </Grid>
          </ContentWrapper>
        </CallbackStateContextProvider>
      </PageTemplate>
    </ThemeProvider>
  );
};
export default AddApplication;
