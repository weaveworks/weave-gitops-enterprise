import React, { FC, useMemo, Dispatch } from 'react';
import { theme as weaveTheme, Button } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import {
  Template,
  TemplateObject,
} from '../../../../../cluster-services/cluster_services.pb';

const base = weaveTheme.spacing.base;
const small = weaveTheme.spacing.small;

const FormWrapper = styled.div`
  .form-group {
    padding-top: ${base};
  }

  .previewCTA {
    display: flex;
    justify-content: flex-end;
    padding-top: ${small};
    padding-bottom: ${base};
  }
`;

const TemplateFields: FC<{
  activeTemplate: Template | null;
  onPRPreview: () => void;
  onFormDataUpdate: Dispatch<React.SetStateAction<any>>;
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
}> = ({ activeTemplate, onPRPreview, formData, setFormData }) => {
  const parameters = useMemo(() => {
    return (
      activeTemplate?.parameters?.map(param => {
        const { options } = param;
        const name = param.name as string;
        if (options?.length !== 0) {
          return {
            [name]: {
              type: 'string',
              title: `${name}`,
              enum: options,
            },
          };
        } else {
          return {
            [name]: {
              type: 'string',
              title: `${name}`,
            },
          };
        }
      }) || []
    );
  }, [activeTemplate]);

  console.log(parameters);

  // }, [sections, addUserSelectedFields, userSelectedFields]);

  // const credentialsItems = [
  //   ...credentials.map((credential: Credential) => {
  //     const { kind, namespace, name } = credential;
  //     return (
  //       <MenuItem key={name} value={name || ''}>
  //         {`${kind}/${namespace || 'default'}/${name}`}
  //       </MenuItem>
  //     );
  //   }),
  //   <MenuItem key="None" value="None">
  //     <em>None</em>
  //   </MenuItem>,
  // ];

  return (
    /* TO DO: if enum show select else show the normal input */

    // formData structure

    // { ... parameter_values":{"url":"https://github.com/wkp-example-org/capd-demo-reloaded","provider":"GitHub","branchName":"create-clusters-branch-ded8a","pullRequestTitle":"Creates capi cluster",""commitMessage":"Creates capi cluster","pullRequestDescription":"This PR creates a new cluster","CLUSTER_NAME":"testali111","NAMESPACE":"default","CONTROL_PLANE_MACHINE_COUNT":"1","KUBERNETES_VERSION":"1.19.11","WORKER_MACHINE_COUNT":"1"} ... ]}

    <FormWrapper>
      {/* <div className="profile-namespace">
        <span>Namespace</span>
        <FormControl>
          <Input
            id="profile-namespace"
            value={namespace}
            placeholder="flux-system"
            onChange={handleChangeNamespace}
            error={!isNamespaceValid}
          />
        </FormControl>
      </div> */}

      {/* <div className="credentials">
        <span>Infrastructure provider credentials:</span>
        <FormControl>
          <Select
            disabled={isLoading}
            value={infraCredential?.name || 'None'}
            onChange={handleSelectCredentials}
            autoWidth
            label="Credentials"
          >
            {credentialsItems}
          </Select>
        </FormControl>
      </div> */}
      <div className="previewCTA">
        <Button type="submit">PREVIEW PR</Button>
      </div>
    </FormWrapper>
  );
};

export default TemplateFields;
