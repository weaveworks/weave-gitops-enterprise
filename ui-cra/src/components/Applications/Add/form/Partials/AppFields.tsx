import React, { FC, Dispatch } from 'react';
import styled from 'styled-components';
import useClusters from '../../../../../contexts/Clusters';
import { Input, Select } from '../../../../../utils/form';
import _ from 'lodash';
import { Loader } from '../../../../Loader';
import { MenuItem } from '@material-ui/core';
import { GitopsClusterEnriched } from '../../../../../types/custom';
import { useListSources } from '@weaveworks/weave-gitops';
import { Source } from '@weaveworks/weave-gitops/ui/lib/types';

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
  setFormData: Dispatch<React.SetStateAction<any>>;
}> = ({ formData, setFormData }) => {
  const { clusters, isLoading } = useClusters();
  const { data } = useListSources();

  console.log(data);

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

  const repos = data?.result.filter(
    object =>
      object.clusterName === clusterName &&
      (object.kind === 'KindGitRepository' ||
        object.kind === 'KindHelmRepository'),
  );

  const handleSelectSource = (event: React.ChangeEvent<any>) => {
    const { value } = event.target;

    setFormData({
      ...formData,
      source_name: JSON.parse(value).name,
      source_namespace: JSON.parse(value).namespace,
      source_type: JSON.parse(value).kind,
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
    <FormWrapper>
      <Input
        className="form-section"
        required={true}
        name="name"
        label="APPLICATION NAME"
        value={formData.name}
        onChange={event => handleFormData(event, 'name')}
      />
      <Input
        className="form-section"
        required={true}
        name="namespace"
        label="APPLICATION NAMESPACE"
        value={formData.namespace}
        onChange={event => handleFormData(event, 'namespace')}
      />
      <div>
        {!isLoading ? (
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
            {clusters?.map((option: GitopsClusterEnriched, index: number) => {
              return (
                <MenuItem key={index} value={JSON.stringify(option)}>
                  {option.name}
                </MenuItem>
              );
            })}
          </Select>
        ) : (
          <div className="loader">
            <Loader />
          </div>
        )}
      </div>
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
        {repos ? (
          repos.map((option: Source, index: number) => {
            return (
              <MenuItem key={index} value={JSON.stringify(option)}>
                {option.name}
              </MenuItem>
            );
          })
        ) : (
          <MenuItem disabled={true}>
            No repositories available, please select another cluster.
          </MenuItem>
        )}
      </Select>
      {formData.source_type === 'KindGitRepository' ? (
        <Input
          className="form-section"
          required={true}
          name="path"
          label="SELECT PATH/CHART"
          value={formData.path}
          onChange={event => handleFormData(event, 'path')}
          description="Path within the git repository to read yaml files"
        />
      ) : null}
    </FormWrapper>
  );
};

export default AppFields;
