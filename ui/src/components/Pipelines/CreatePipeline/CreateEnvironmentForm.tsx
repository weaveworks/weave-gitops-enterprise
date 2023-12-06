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

interface Environment {
  name: string;
  strategy: string;
  branch: string;
  credentialSecret: string;
  repoType: string;
  repoUrl: string;
  manualApproval: boolean;
}
interface CreateEnvironmentProps {
  environment?: Environment;
}

const initialFormData: Environment = {
  name: '',
  strategy: '',
  branch: '',
  credentialSecret: '',
  repoType: '',
  repoUrl: '',
  manualApproval: false,
};

const CreateEnvironmentFormWrapper = styled(Flex)`
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

const CreateEnvironmentForm = ({ environment }: CreateEnvironmentProps) => {
  const [formData, setFormData] = useState<Environment>(
    environment || initialFormData,
  );
  const handleFormData = (value: any, key: string) => {
    setFormData(f => ({ ...f, [key]: value }));
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
          value={formData.strategy}
          onChange={event => {
            handleFormData(event.target.value, 'strategy');
          }}
        >
          <FormControlLabel
            value={'pull request'}
            control={<Radio />}
            label="Pull request"
          />
          <FormControlLabel
            value={'notification'}
            control={<Radio />}
            label="Notification"
          />
        </RadioGroup>
        <FormControlLabel
          control={
            <Checkbox
              checked={formData.manualApproval}
              color="primary"
              onChange={event => {
                handleFormData(event.target.value, 'manualApproval');
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
          value={formData.branch}
          handleFormData={val => handleFormData(val, 'branch')}
          // error={!formData.branch}
        />
        <InputDebounced
          // required
          name="credentialSecret"
          label="GIT CREDENTIAL SECRET"
          value={formData.credentialSecret}
          handleFormData={val => handleFormData(val, 'credentialSecret')}
          // error={!formData.credentialSecret}
        />

        <Select
          value={formData.repoType}
          onChange={event => {
            handleFormData(event.target.value, 'repoType');
          }}
          label="APP REPO TYPE"
          items={['Github', 'Gitlab', 'Bit Bucket']}
        />
        <InputDebounced
          // required
          name="repoUrl"
          label="APP REPO URL"
          value={formData.repoUrl}
          handleFormData={val => handleFormData(val, 'repoUrl')}
          // error={!formData.repoUrl}
        />
      </Flex>
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
