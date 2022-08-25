import React, { FC, Dispatch } from 'react';
import styled from 'styled-components';
import { Input, Select } from '../../../../../utils/form';
import _ from 'lodash';
import { MenuItem } from '@material-ui/core';
import { GitRepository } from '@weaveworks/weave-gitops/ui/lib/api/core/types.pb';
import { GitopsClusterEnriched } from '../../../../../types/custom';
import { ListGitRepositoriesResponse } from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';
import { DEFAULT_FLUX_KUSTOMIZATION_NAMESPACE } from '../../../../../utils/config';

const FormWrapper = styled.form`
  .form-section {
    width: 50%;
  }
  .loader {
    padding-bottom: ${({ theme }) => theme.spacing.medium};
  }
`;

const AppFields: FC<{
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>> | any;
  index?: number;
  clusters?: GitopsClusterEnriched[];
  GitRepoResponse?: ListGitRepositoriesResponse;
}> = ({
  formData,
  setFormData,
  index = 0,
  clusters = undefined,
  GitRepoResponse = undefined,
}) => {
  const automation = formData.clusterAutomations[index];
  let gitResposFilterdList: GitRepository[] = [];

  const handleSelectCluster = (event: React.ChangeEvent<any>) => {
    const value = event.target.value;
    let currentAutomation = [...formData.clusterAutomations];
    currentAutomation[index] = {
      ...automation,
      cluster_name: JSON.parse(value).name,
      cluster_namespace: JSON.parse(value).namespace,
      cluster_isControlPlane: JSON.parse(value).controlPlane,
      cluster: value,
    };
    setFormData({
      ...formData,
      clusterAutomations: currentAutomation,
    });
  };
  if (clusters) {
    const clusterName = automation.cluster_namespace
      ? `${automation.cluster_namespace}/${automation.cluster_name}`
      : `${automation.cluster_name}`;
    gitResposFilterdList = _.filter(GitRepoResponse?.gitRepositories, [
      'clusterName',
      clusterName,
    ]);
  }

  const handleSelectSource = (event: React.ChangeEvent<any>) => {
    const value = event.target.value;
    let currentAutomation = [...formData.clusterAutomations];
    currentAutomation[index] = {
      ...automation,
      source_name: JSON.parse(value).name,
      source_namespace: JSON.parse(value).namespace,
      source: value,
    };
    setFormData({
      ...formData,
      clusterAutomations: currentAutomation,
    });
  };

  const handleFormData = (
    event:
      | React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>
      | React.ChangeEvent<{ name?: string; value: unknown }>,
    fieldName?: string,
  ) => {
    const { value } = event?.target;

    let currentAutomation = [...formData.clusterAutomations];
    currentAutomation[index] = {
      ...automation,
      [fieldName as string]: value,
    };
    setFormData({
      ...formData,
      clusterAutomations: currentAutomation,
    });
  };

  return (
    <FormWrapper>
      <Input
        className="form-section"
        required={true}
        name="name"
        label="KUSTOMIZATION NAME"
        value={formData.clusterAutomations[index].name}
        onChange={event => handleFormData(event, 'name')}
      />
      <Input
        className="form-section"
        name="namespace"
        label="KUSTOMIZATION NAMESPACE"
        placeholder={DEFAULT_FLUX_KUSTOMIZATION_NAMESPACE}
        value={formData.clusterAutomations[index].namespace}
        onChange={event => handleFormData(event, 'namespace')}
      />
      {!!clusters && (
        <>
          <Select
            className="form-section"
            name="cluster_name"
            required={true}
            label="SELECT CLUSTER"
            value={formData.clusterAutomations[index].cluster || ''}
            onChange={handleSelectCluster}
            defaultValue={''}
            description="select target cluster"
          >
            {clusters?.map((option: GitopsClusterEnriched, index: number) => {
              return (
                <MenuItem key={index} value={JSON.stringify(option)}>
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
            value={formData.clusterAutomations[index].source || ''}
            onChange={handleSelectSource}
            defaultValue={''}
            description="The name and type of source"
          >
            {gitResposFilterdList.length > 0 ? (
              gitResposFilterdList?.map(
                (option: GitRepository, index: number) => {
                  return (
                    <MenuItem key={index} value={JSON.stringify(option)}>
                      {option.name}
                    </MenuItem>
                  );
                },
              )
            ) : (
              <MenuItem disabled={true}>
                No GitRepository available please select another cluster
              </MenuItem>
            )}
          </Select>
        </>
      )}
      <Input
        className="form-section"
        required={true}
        name="path"
        label="SELECT PATH/CHART"
        value={formData.clusterAutomations[index].path}
        onChange={event => handleFormData(event, 'path')}
        description="Path within the git repository to read yaml files"
      />
    </FormWrapper>
  );
};

export default AppFields;
