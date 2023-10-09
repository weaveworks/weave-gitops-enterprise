import { MenuItem } from '@material-ui/core';
import { Flex, Kind } from '@weaveworks/weave-gitops';
import { useState } from 'react';
import styled from 'styled-components';
import { ExternalSecretItem } from '../../../cluster-services/cluster_services.pb';
import CallbackStateContextProvider from '../../../contexts/GitAuth/CallbackStateContext';
import { Input, Select } from '../../../utils/form';
import { Routes } from '../../../utils/nav';
import { Page } from '../../Layout/App';
import { NotificationsWrapper } from '../../Layout/NotificationsWrapper';
import GitOps from '../../Templates/Form/Partials/GitOps';
import { FormWrapper } from '../../Templates/Form/utils';
import SelectSecret from './form/Partials/SelectSecret';

type Props = {
  className?: string;
};

function AddSource({ className }: Props) {
  const supportedSourceKinds = [
    Kind.GitRepository,
    Kind.HelmRepository,
    Kind.Bucket,
    Kind.OCIRepository,
  ];

  const [showAuthDialog, setShowAuthDialog] = useState(false);
  const [kind, setKind] = useState<Kind>(Kind.GitRepository);
  const [commonData, setCommonData] = useState({
    name: '',
    namespace: 'flux-system',
    interval: '1m0s',
  });
  const [kindOptions, setKindOptions] = useState({
    url: '',
    tag: '',
  });
  const [gitopsData, setGitopsData] = useState({
    branch: '',
    message: '',
    prTitle: '',
    prDesc: '',
  });
  const [secret, setSecret] = useState<ExternalSecretItem>({});
  // const [openPreview, setOpenPreview] = useState(false);
  return (
    <Page
      path={[
        {
          label: 'Sources',
          url: Routes.Sources,
        },
        { label: 'Add new source' },
      ]}
    >
      <CallbackStateContextProvider
        callbackState={{
          page: Routes.AddSource,
          state: {
            //   formData,
            //   updatedProfiles,
          },
        }}
      >
        <NotificationsWrapper>
          <FormWrapper
            noValidate
            //   onSubmit={event =>
            // validateFormData(
            //   event,
            //   submitType === 'PR Preview'
            //     ? handlePRPreview
            //     : handleAddApplication,
            //   setFormError,
            //   setSubmitType,
            // )
            //   }
          >
            <Select
              name="source-kind"
              required={true}
              label="SELECT KIND"
              value={kind}
              onChange={e => setKind(e.target.value as Kind)}
              defaultValue={Kind.GitRepository}
            >
              {supportedSourceKinds.map((kind: Kind, index: number) => {
                return (
                  <MenuItem key={index} value={kind}>
                    {kind}
                  </MenuItem>
                );
              })}
            </Select>
            <Flex wide gap="12" start>
              <Input
                required={true}
                label="NAME"
                value={commonData.name}
                onChange={event =>
                  setCommonData({ ...commonData, name: event.target.value })
                }
              />
              <Input
                label="NAMESPACE"
                value={commonData.namespace}
                onChange={event =>
                  setCommonData({
                    ...commonData,
                    namespace: event.target.value,
                  })
                }
              />
              <Input
                label="INTERVAL"
                value={commonData.interval}
                onChange={event =>
                  setCommonData({
                    ...commonData,
                    interval: event.target.value,
                  })
                }
              />
            </Flex>
            <Flex wide gap="12" wrap>
              {kind !== Kind.Bucket && (
                <Input
                  label="URL"
                  value={kindOptions.url}
                  onChange={event =>
                    setKindOptions({
                      ...kindOptions,
                      url: event.target.value,
                    })
                  }
                />
              )}
            </Flex>
            <SelectSecret secret={secret} setSecret={setSecret} />
            {/* <Preview openPreview={openPreview} setOpenPreview={setOpenPreview}/> */}
            <GitOps
              formData={gitopsData}
              setFormData={setGitopsData}
              showAuthDialog={showAuthDialog}
              setShowAuthDialog={setShowAuthDialog}
              formError={''}
              enableGitRepoSelection={true}
            />
          </FormWrapper>
        </NotificationsWrapper>
      </CallbackStateContextProvider>
    </Page>
  );
}

export default styled(AddSource).attrs({ className: AddSource.name })``;
