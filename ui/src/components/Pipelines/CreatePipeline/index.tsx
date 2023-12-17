import {
  Button,
  Flex,
  Icon,
  IconType,
  Page,
  Text,
} from '@weaveworks/weave-gitops';
import { useCallback, useMemo, useState } from 'react';
import styled from 'styled-components';
import { GitProvider } from '../../../api/gitauth/gitauth.pb';
import CallbackStateContextProvider from '../../../contexts/GitAuth/CallbackStateContext';
import useNotifications from '../../../contexts/Notifications';
import {
  expiredTokenNotification,
  useIsAuthenticated,
} from '../../../hooks/gitprovider';
import { useCallbackState } from '../../../utils/callback-state';
import { InputDebounced, validateFormData } from '../../../utils/form';
import { Routes } from '../../../utils/nav';
import { removeToken } from '../../../utils/request';
import { clearCallbackState, getProviderToken } from '../../GitAuth/utils';
import { NotificationsWrapper } from '../../Layout/NotificationsWrapper';
import { handleError, scrollToAlertSection } from '../../Secrets/Shared/utils';
import GitOps from '../../Templates/Form/Partials/GitOps';
import { FormWrapper } from '../../Templates/Form/utils';
import CreateEnvironmentForm from './CreateEnvironmentForm';
import EnvironmentCard from './EnvironmentCard';
import ListApplications from './ListApplications';
import { Pipeline, getPipelineInitialData } from './utils';

const PipelineContainer = styled(Flex)`
  border-bottom: 1px solid ${props => props.theme.colors.neutral20};

  .MuiInputBase-root {
    margin-right: 0;
  }
`;

const CreateEnvironmentContainer = styled(Flex)`
  .w-50 {
    width: 50%;
  }
  .card {
    padding: 20px;
    background: ${props => props.theme.colors.pipelinesBackGray};
    height: 595px;
    border-radius: 4px;
    overflow: auto;
  }
`;

const CreatePipeline = () => {
  const callbackState = useCallbackState();
  const random = useMemo(() => Math.random().toString(36).substring(7), []);
  const { initialFormData } = getPipelineInitialData(callbackState, random);

  const [showAuthDialog, setShowAuthDialog] = useState(false);

  const [formError, setFormError] = useState<string>('');
  const [validateForm, setValidateForm] = useState<boolean>(false);
  const [formData, setFormData] = useState<Pipeline>(initialFormData);
  const handleFormData = (value: any, key: string) => {
    setFormData(f => ({ ...f, [key]: value }));
  };

  const { setNotifications } = useNotifications();

  const [loading, setLoading] = useState<boolean>(false);
  const token = getProviderToken(formData.provider as GitProvider);

  const { isAuthenticated, validateToken } = useIsAuthenticated(
    formData.provider as GitProvider,
    token,
  );

  const handleCreatePipeline = useCallback(() => {
    setLoading(true);

    validateToken()
      .then(async () => {
        try {
          //   const payload = getESFormattedPayload(formData);
          //   const response = await api.CreateAutomationsPullRequest(
          //     {
          //       headBranch: formData.branchName,
          //       title: formData.pullRequestTitle,
          //       description: formData.pullRequestDescription,
          //       commitMessage: formData.commitMessage,
          //       repositoryUrl: getRepositoryUrl(formData.repo as GitRepository),
          //       clusterAutomations: [payload],
          //     },
          //     {
          //       headers: new Headers({ 'Git-Provider-Token': `token ${token}` }),
          //     },
          //   );
          //   setNotifications([
          //     {
          //       message: {
          //         component: (
          //           <Link href={response.webUrl} newTab>
          //             PR created successfully, please review and merge the pull
          //             request to apply the changes to the cluster.
          //           </Link>
          //         ),
          //       },
          //       severity: 'success',
          //     },
          //   ]);
          scrollToAlertSection();
        } catch (error: any) {
          handleError(error, setNotifications);
        } finally {
          setLoading(false);
          removeToken(formData.provider);
          clearCallbackState();
        }
      })
      .catch(() => {
        removeToken(formData.provider);
        setNotifications([expiredTokenNotification]);
      })
      .finally(() => setLoading(false));
  }, [formData.provider, setNotifications, validateToken]);

  const authRedirectPage = Routes.CreateSecret;

  return (
    <Page
      path={[
        { label: 'Pipeline', url: Routes.Pipelines },
        { label: 'Create and Edit Your Pipeline' },
      ]}
    >
      <CallbackStateContextProvider
        callbackState={{
          page: authRedirectPage,
          state: {
            formData,
          },
        }}
      >
        <NotificationsWrapper>
          <FormWrapper
            noValidate
            onSubmit={event => {
              setValidateForm(true);
              validateFormData(event, handleCreatePipeline, setFormError);
            }}
          >
            <Flex wide column gap="24">
              <PipelineContainer wide column gap="20">
                <Flex column gap="4">
                  <Text color="neutral30" size="large">
                    Use this area to set up your pipeline or to make changes.
                  </Text>
                  <Text color="neutral30" size="large">
                    When you're done, click the "Apply" button at the bottom.
                  </Text>
                </Flex>
                <Flex wide center alignItems="start" between gap="20">
                  <InputDebounced
                    required
                    name="pipelineName"
                    label="ENTER NEW PIPELINE NAME"
                    value={formData.pipelineName}
                    handleFormData={val => handleFormData(val, 'pipelineName')}
                    error={validateForm && !formData.pipelineName}
                  />
                  <InputDebounced
                    required
                    name="namespace"
                    label="ADD NAMESPACE"
                    value={formData.pipelineNamespace}
                    handleFormData={val =>
                      handleFormData(val, 'pipelineNamespace')
                    }
                    error={validateForm && !formData.pipelineNamespace}
                  />
                  <ListApplications
                    value={formData.applicationName}
                    handleFormData={val =>
                      handleFormData(val, 'applicationName')
                    }
                    validateForm={validateForm && !formData.applicationName}
                  />
                </Flex>
              </PipelineContainer>

              <Flex wide column center gap="20">
                <Flex wrap between wide>
                  <Flex column gap="4">
                    <Text color="neutral30" size="large">
                      Add a new environment to your pipeline
                    </Text>
                    <Text color="neutral30">
                      Please add your environments in the order you choose for
                      your pipeline.
                    </Text>
                  </Flex>
                  <Flex gap="16">
                    <Button
                      id="create-environment"
                      startIcon={<Icon type={IconType.EditIcon} size="base" />}
                      onClick={() => {}}
                    >
                      CREATE ENVIRONMENT
                    </Button>
                    <Button
                      id="create-environment"
                      startIcon={<Icon type={IconType.AddIcon} size="base" />}
                      onClick={() => {}}
                    >
                      ADD TARGET
                    </Button>
                  </Flex>
                </Flex>
                <CreateEnvironmentContainer wide gap="20">
                  <div className="card w-50">
                    {/* <FollowSteps /> */}
                    <CreateEnvironmentForm />
                    {/* <AddTarget /> */}
                  </div>
                  <div className="card w-50">
                    <EnvironmentCard />
                  </div>
                </CreateEnvironmentContainer>
              </Flex>
              <GitOps
                loading={loading}
                isAuthenticated={isAuthenticated}
                formData={formData}
                setFormData={setFormData}
                showAuthDialog={showAuthDialog}
                setShowAuthDialog={setShowAuthDialog}
                formError={formError}
                enableGitRepoSelection={true}
              >
                {/* <Preview formData={formData} setFormError={setFormError} /> */}
              </GitOps>
            </Flex>
          </FormWrapper>
        </NotificationsWrapper>
      </CallbackStateContextProvider>
    </Page>
  );
};

export default CreatePipeline;
