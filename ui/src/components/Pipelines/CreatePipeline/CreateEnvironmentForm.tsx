import {
  Checkbox,
  FormControlLabel,
  Radio,
  RadioGroup,
} from '@material-ui/core';
import { Button, Flex, Icon, IconType, Text } from '@weaveworks/weave-gitops';
import { useState } from 'react';
import styled from 'styled-components';
import { InputDebounced, Select } from '../../../utils/form';
import ListClusters from '../../Secrets/Shared/ListClusters';

interface Environment {
  name: string;
  manualPromotion: boolean;

  pullRequestType?: string;
  pullRequestUrl?: string;
  pullRequestBranch?: string;
  pullRequestCredentialSecret?: string;

  notification: boolean;

  targets: {
    namespace?: string;
    clusterName?: string;
    clusterNamespace?: string;
  }[];
}
export interface TargetProps {
  clusterName?: string;
  namespace?: string;
  clusterNamespace?: string;
  environment?: string;
}
interface CreateEnvironmentProps {
  environment?: Environment;
}

const initialFormData: Environment = {
  name: '',
  targets: [],
  manualPromotion: false,
  notification: false
};

export const CreateEnvironmentFormWrapper = styled(Flex)`
  background: ${props => props.theme.colors.white};
  padding: 24px;
  border-radius: 8px;

  div[class*='MuiFormControl-root'] {
    width: 100%;
    padding-bottom: 0px;
  }
  .MuiInputBase-root {
    margin-right: 0;
  }
  .MuiFormControlLabel-root {
    margin: 0px;
  }
`;

const AddTarget = ({
  addTarget,
}: {
  addTarget: (newTarget: TargetProps) => void;
}) => {
  const [targetForm, setTarget] = useState<TargetProps>({});

  const handleAddTarget = () => {
    addTarget(targetForm);
    setTarget({});
  };

  return (
    <>
      <ListClusters
        value={targetForm?.clusterName || ''}
        validateForm={false}
        handleFormData={(val: any) => {
          const [ns, cluster] = val.split('/');

          setTarget(f => ({
            ...f,
            clusterName: !cluster ? ns : cluster,
            clusterNamespace: cluster ? ns : '',
            namespace: '',
          }));
        }}
      />
      <InputDebounced
        required
        name="namespace"
        label="Namespace"
        value={targetForm.namespace || ''}
        handleFormData={val => {
          setTarget(f => ({
            ...f,
            namespace: val,
          }));
        }}
      />
      <Button onClick={handleAddTarget}>Add Target</Button>
    </>
  );
};


const CreateEnvironmentForm = ({ environment }: CreateEnvironmentProps) => {
  const [formData, setFormData] = useState<Environment>(
    environment || initialFormData,
  );

  const [strategy, setStrategy] = useState('pull_request');

  const handleFormData = (value: any, key: string) => {
    setFormData(f => ({ ...f, [key]: value }));
  };
  const handleAddTarget = (newTarget: TargetProps) => {
    setFormData((f) => ({
      ...f,
      targets: [...f.targets, newTarget],
    }));
  };
  const handleCreateEnvironment = () => {};

  return (
    <CreateEnvironmentFormWrapper column gap="16">
      <Text semiBold>Create an Environment</Text>
      <InputDebounced
        // required
        name="name"
        label="NAME"
        value={formData.name}
        handleFormData={val => handleFormData(val, 'name')}
        // error={!formData.name}
      />

      <Text color="neutral30">CHOOSE PROMOTION STRATEGY</Text>
      <Flex between wide>
        <RadioGroup
          row
          name="row-radio-buttons-group"
          value={strategy}
          onChange={event => {
            setStrategy(event.target.value);
          }}
        >
          <FormControlLabel
            value="pull_request"
            control={<Radio />}
            label="Pull request"
          />
          <FormControlLabel
            value="notification"
            control={<Radio />}
            label="Notification"
          />
        </RadioGroup>
        <FormControlLabel
          control={
            <Checkbox
              checked={formData.manualPromotion}
              color="primary"
              onChange={event => {
                handleFormData(event.target.value, 'manualPromotion');
              }}
            />
          }
          label="Apply manual approval"
        />
      </Flex>

      <Flex column gap="16" wide>
        <InputDebounced
          // required
          name="branch"
          label="APP REPO BRANCH"
          description="Branch to write updates to in Git"
          value={formData.pullRequestBranch}
          handleFormData={val => handleFormData(val, 'pullRequestBranch')}
          // error={!formData.branch}
        />
        <InputDebounced
          // required
          name="credentialSecret"
          label="GIT CREDENTIAL SECRET"
          value={formData.pullRequestCredentialSecret}
          handleFormData={val =>
            handleFormData(val, 'pullRequestCredentialSecret')
          }
          // error={!formData.credentialSecret}
        />

        <Select
          value={formData.pullRequestType || ''}
          onChange={event => {
            handleFormData(event.target.value, 'pullRequestType');
          }}
          label="APP REPO TYPE"
          items={['Github', 'Gitlab', 'Bit Bucket']}
        />
        <InputDebounced
          // required
          name="repoUrl"
          label="APP REPO URL"
          value={formData.pullRequestUrl}
          handleFormData={val => handleFormData(val, 'pullRequestUrl')}
          // error={!formData.repoUrl}
        />
      </Flex>

      <Text semiBold>Targets:</Text>
      {formData.targets.map((target, index) => (
        <div key={index}>
          {`Cluster: ${target.clusterName}, Namespace: ${target.namespace}`}
        </div>
      ))}

      <Text semiBold>Target form:</Text>
      <AddTarget addTarget={handleAddTarget} />

      <Button
        id="create-policy-config"
        startIcon={<Icon type={IconType.AddIcon} size="base" />}
        onClick={handleCreateEnvironment}
      >
        APPLY
      </Button>
    </CreateEnvironmentFormWrapper>
  );
};

export default CreateEnvironmentForm;
