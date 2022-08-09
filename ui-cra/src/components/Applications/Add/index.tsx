import React, { useCallback, useEffect, useState } from 'react';
import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { useApplicationsCount } from '../utils';
import useClusters from '../../../contexts/Clusters';
import GitOps from '../../Clusters/Create/Form/Partials/GitOps';
import { Grid } from '@material-ui/core';
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

const AddApplication = () => {
  const applicationsCount = useApplicationsCount();
  const { clusters, isLoading } = useClusters();
  const [loading, setLoading] = useState<boolean>(false);
  const [showAuthDialog, setShowAuthDialog] = useState(false);
  const { addApplication } = useTemplates();
  const [PRPreview, setPRPreview] = useState<string | null>(null);
  const history = useHistory();
  const { setNotifications } = useNotifications();
  const { data } = useListConfig();
  const repositoryURL = data?.repositoryURL || '';
  const authRedirectPage = `/applications/new`;

  let initialFormData = {
    url: '',
    provider: '',
    branchName: `add-application-branch`,
    pullRequestTitle: 'Add application',
    commitMessage: 'add application',
    pullRequestDescription: 'This PR add a new application',
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

  const handleAddApplication = useCallback(() => {
    const payload = {
      head_branch: formData.branchName,
      title: formData.pullRequestTitle,
      description: formData.pullRequestDescription,
      commit_message: formData.commitMessage,
    };
    setLoading(true);
    return addApplication(
      payload,
      getProviderToken(formData.provider as GitProvider),
    )
      .then(response => {
        setPRPreview(null);
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
  }, [addApplication, formData, history, setNotifications, setPRPreview]);

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
