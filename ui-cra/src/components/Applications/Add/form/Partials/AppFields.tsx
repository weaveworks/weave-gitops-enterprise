import React, { FC, Dispatch } from 'react';
import styled from 'styled-components';
import { Input, Select } from '../../../../../utils/form';
import _ from 'lodash';
import { MenuItem } from '@material-ui/core';
import { GitRepository } from '@weaveworks/weave-gitops/ui/lib/api/core/types.pb';
import { GitopsClusterEnriched } from '../../../../../types/custom';
import { ListGitRepositoriesResponse } from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';

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
  let gitResposFilterdList: GitRepository[] = [];

  const handleSelectCluster = (event: React.ChangeEvent<any>) => {
    const value = event.target.value;
    let currentAutomation = [...formData.clusterAutomations];
    currentAutomation[index] = {
      ...currentAutomation[index],
      cluster_name: JSON.parse(value).name,
      cluster_namespace: JSON.parse(value).namespace,
      cluster_isControlPlane: JSON.parse(value).controlPlane,
      cluster: value,
    };
    setFormData({
      ...formData,
      clusterAutomations: [...currentAutomation],
    });
  };
  if (clusters) {
    const clusterName = formData.clusterAutomations[0].cluster_namespace
      ? `${formData.clusterAutomations[0].cluster_namespace}/${formData.clusterAutomations[0].cluster_name}`
      : `${formData.clusterAutomations[0].cluster_name}`;
    gitResposFilterdList = _.filter(GitRepoResponse?.gitRepositories, [
      'clusterName',
      clusterName,
    ]);
  }

  const handleSelectSource = (event: React.ChangeEvent<any>) => {
    const value = event.target.value;
    let currentAutomation = [...formData.clusterAutomations];
    currentAutomation[index] = {
      ...currentAutomation[index],
      source_name: JSON.parse(value).name,
      source_namespace: JSON.parse(value).namespace,
      source: value,
    };
    setFormData({
      ...formData,
      clusterAutomations: [...currentAutomation],
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
      ...currentAutomation[index],
      [fieldName as string]: value,
    };
    setFormData({
      ...formData,
      clusterAutomations: [...currentAutomation],
    });
  };

  return (
    <FormWrapper>
      <Input
        className="form-section"
        required={true}
        name="name"
        label="APPLICATION NAME"
        value={formData.clusterAutomations[index].name}
        onChange={event => handleFormData(event, 'name')}
      />
      <Input
        className="form-section"
        required={true}
        name="namespace"
        label="APPLICATION NAMESPACE"
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
