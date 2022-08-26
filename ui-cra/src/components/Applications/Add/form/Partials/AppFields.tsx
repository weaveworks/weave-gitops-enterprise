import React, { FC, Dispatch } from 'react';
import styled from 'styled-components';
import _ from 'lodash';
import useProfiles from '../../../../../contexts/Profiles';
import { Input, Select } from '../../../../../utils/form';
import { ListSubheader, MenuItem } from '@material-ui/core';
import { GitopsClusterEnriched } from '../../../../../types/custom';
import { useListSources, theme } from '@weaveworks/weave-gitops';
import { DEFAULT_FLUX_KUSTOMIZATION_NAMESPACE } from '../../../../../utils/config';
import { Source } from '@weaveworks/weave-gitops/ui/lib/types';
import { getGitRepoHTTPSURL } from '../../../../../utils/formatters';
import { isAllowedLink } from '@weaveworks/weave-gitops';

interface SourceEnriched extends Source {
  url?: string;
  reference?: {
    branch?: string;
  };
}

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
}> = ({ formData, setFormData, index = 0, clusters = undefined }) => {
  const { setHelmRepo } = useProfiles();
  const { data } = useListSources();
  const automation = formData.clusterAutomations[index];

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

  let gitRepos = [] as Source[];
  let helmRepos = [] as Source[];

  if (clusters) {
    const clusterName = automation.cluster_namespace
      ? `${automation.cluster_namespace}/${automation.cluster_name}`
      : `${automation.cluster_name}`;

    gitRepos = _.orderBy(
      _.filter(
        data?.result,
        item =>
          item.kind === 'KindGitRepository' && item.clusterName === clusterName,
      ),
      ['name'],
      ['asc'],
    );

    helmRepos = _.orderBy(
      _.filter(
        data?.result,
        item =>
          item.kind === 'KindHelmRepository' &&
          item.clusterName === clusterName,
      ),
      ['name'],
      ['asc'],
    );
  }

  const handleSelectSource = (event: React.ChangeEvent<any>) => {
    const { value } = event.target;

    let currentAutomation = [...formData.clusterAutomations];

    currentAutomation[index] = {
      ...automation,
      source_name: JSON.parse(value).name,
      source_namespace: JSON.parse(value).namespace,
      source: value,
    };

    setFormData({
      ...formData,
      source_name: JSON.parse(value).name,
      source_namespace: JSON.parse(value).namespace,
      source_type: JSON.parse(value).kind,
      source: value,
      clusterAutomations: currentAutomation,
    });

    if (JSON.parse(value).kind === 'KindHelmRepository') {
      setHelmRepo({
        name: JSON.parse(value).name,
        namespace: JSON.parse(value).namespace,
      });
    }
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

  const optionUrl = (url?: string, branch?: string) => {
    if (branch) {
      return isAllowedLink(getGitRepoHTTPSURL(url, branch)) ? (
        <a
          title="Visit repository"
          style={{ color: theme.colors.primary, fontSize: '10px' }}
          href={getGitRepoHTTPSURL(url, branch)}
          target="_blank"
          rel="noopener noreferrer"
        >
          {url}@<strong>{branch}</strong>
        </a>
      ) : (
        <span>
          {url}@<strong>{branch}</strong>
        </span>
      );
    } else {
      return isAllowedLink(getGitRepoHTTPSURL(url)) ? (
        <a
          title="Visit repository"
          style={{ color: theme.colors.primary, fontSize: '10px' }}
          href={getGitRepoHTTPSURL(url)}
          target="_blank"
          rel="noopener noreferrer"
        >
          {url}
        </a>
      ) : (
        <span>{url}</span>
      );
    }
  };

  return (
    <FormWrapper>
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
            {[...gitRepos, ...helmRepos].length === 0 && (
              <MenuItem disabled={true}>
                No repository available, please select another cluster.
              </MenuItem>
            )}
            {gitRepos.length !== 0 && (
              <ListSubheader>GitRepository</ListSubheader>
            )}
            {gitRepos?.map((option: SourceEnriched, index: number) => (
              <MenuItem key={index} value={JSON.stringify(option)}>
                {option.name}&nbsp;&nbsp;
                {optionUrl(option?.url, option?.reference?.branch)}
              </MenuItem>
            ))}
            {helmRepos.length !== 0 && (
              <ListSubheader>HelmRepository</ListSubheader>
            )}
            {helmRepos?.map((option: SourceEnriched, index: number) => (
              <MenuItem key={index} value={JSON.stringify(option)}>
                {option.name}&nbsp;&nbsp;
                {optionUrl(option?.url)}
              </MenuItem>
            ))}
          </Select>
        </>
      )}
      {formData.source_type === 'KindGitRepository' ? (
        <>
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
          <Input
            className="form-section"
            required={true}
            name="path"
            label="SELECT PATH/CHART"
            value={formData.clusterAutomations[index].path}
            onChange={event => handleFormData(event, 'path')}
            description="Path within the git repository to read yaml files"
          />
        </>
      ) : null}
    </FormWrapper>
  );
};

export default AppFields;
