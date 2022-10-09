import React, { FC, Dispatch, useEffect, useCallback } from 'react';
import styled from 'styled-components';
import _ from 'lodash';
import useProfiles from '../../../../../contexts/Profiles';
import { Input, Select } from '../../../../../utils/form';
import {
  ListSubheader,
  MenuItem,
  Checkbox,
  FormControlLabel,
} from '@material-ui/core';
import { useListSources, theme, Flex, Link } from '@weaveworks/weave-gitops';
import { DEFAULT_FLUX_KUSTOMIZATION_NAMESPACE } from '../../../../../utils/config';
import {
  GitRepository,
  HelmRepository,
} from '@weaveworks/weave-gitops/ui/lib/objects';
import { Kind } from '@weaveworks/weave-gitops';
import { getGitRepoHTTPSURL } from '../../../../../utils/formatters';
import { isAllowedLink } from '@weaveworks/weave-gitops';
import { Tooltip } from '../../../../Shared';
import { GitopsCluster } from '../../../../../cluster-services/cluster_services.pb';
import { useClustersWithSources } from '../../../utils';
import { useHistory, useLocation } from 'react-router-dom';

const FormWrapper = styled.form`
  .form-section {
    width: 50%;
  }
  .loader {
    padding-bottom: ${({ theme }) => theme.spacing.medium};
  }
  .input-wrapper {
    padding-bottom: ${({ theme }) => theme.spacing.medium};
  }
  .preview-cta {
    display: flex;
    justify-content: flex-end;
    padding: ${({ theme }) => theme.spacing.small}
      ${({ theme }) => theme.spacing.base};
    button {
      width: 200px;
    }
  }
  .preview-loading {
    padding: ${({ theme }) => theme.spacing.base};
  }
`;

const AppFields: FC<{
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>> | any;
  index?: number;
  onPRPreview?: () => void;
  previewLoading?: boolean;
  allowSelectCluster: boolean;
  context?: string;
  clusterName?: string;
}> = ({
  formData,
  setFormData,
  index = 0,
  allowSelectCluster,
  clusterName,
}) => {
  const { setHelmRepo } = useProfiles();
  const { data } = useListSources();
  const automation = formData.clusterAutomations[index];
  const { createNamespace } = automation;
  const history = useHistory();
  const location = useLocation();

  let clusters: GitopsCluster[] | undefined =
    useClustersWithSources(allowSelectCluster);

  const updateCluster = useCallback(
    (cluster: GitopsCluster) => {
      setFormData((formData: any) => {
        const params = new URLSearchParams(`clusterName=${cluster.name}`);
        history.replace({
          pathname: location.pathname,
          search: params.toString(),
        });
        let currentAutomation = [...formData.clusterAutomations];
        currentAutomation[index] = {
          ...currentAutomation[index],
          cluster_name: cluster.name,
          cluster_namespace: cluster.namespace,
          cluster_isControlPlane: cluster.controlPlane,
          cluster: JSON.stringify(cluster),
        };
        return {
          ...formData,
          clusterAutomations: currentAutomation,
        };
      });
    },
    [index, setFormData, history, location.pathname],
  );

  useEffect(() => {
    if (clusterName && clusters) {
      const cluster = clusters.find(
        (c: GitopsCluster) => c.name === clusterName,
      );
      if (cluster) {
        updateCluster(cluster);
      }
    }
  }, [clusterName, clusters, updateCluster]);

  const handleSelectCluster = (event: React.ChangeEvent<any>) => {
    const value = event.target.value;
    updateCluster(JSON.parse(value));
  };

  let gitRepos: GitRepository[] = [];
  let helmRepos: HelmRepository[] = [];

  if (clusters) {
    const clusterName = automation.cluster_namespace
      ? `${automation.cluster_namespace}/${automation.cluster_name}`
      : `${automation.cluster_name}`;

    gitRepos = _.orderBy(
      _.filter(
        data?.result,
        (item): item is GitRepository =>
          item.type === Kind.GitRepository && item.clusterName === clusterName,
      ),
      ['name'],
      ['asc'],
    );

    helmRepos = _.orderBy(
      _.filter(
        data?.result,
        (item): item is HelmRepository =>
          item.type === Kind.HelmRepository && item.clusterName === clusterName,
      ),
      ['name'],
      ['asc'],
    );
  }

  const handleSelectSource = (event: React.ChangeEvent<any>) => {
    const { value } = event.target;
    const { obj } = JSON.parse(value);

    let currentAutomation = [...formData.clusterAutomations];

    currentAutomation[index] = {
      ...automation,
      source_name: obj?.metadata.name,
      source_namespace: obj?.metadata?.namespace,
      source: value,
    };

    setFormData({
      ...formData,
      source_name: obj?.metadata?.name,
      source_namespace: obj?.metadata?.namespace,
      source_type: obj?.kind,
      source: value,
      clusterAutomations: currentAutomation,
    });

    if (obj?.kind === 'HelmRepository') {
      setHelmRepo({
        name: obj?.metadata?.name,
        namespace: obj?.metadata?.namespace,
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

  const handleCreateNamespace = (
    event: React.ChangeEvent<HTMLInputElement>,
  ) => {
    let currentAutomation = [...formData.clusterAutomations];

    currentAutomation[index] = {
      ...automation,
      createNamespace: event.target.checked,
    };

    setFormData({
      ...formData,
      clusterAutomations: currentAutomation,
    });
  };

  const optionUrl = (url?: string, branch?: string) => {
    const linkText = branch ? (
      <>
        {url}@<strong>{branch}</strong>
      </>
    ) : (
      url
    );
    if (branch) {
      return isAllowedLink(getGitRepoHTTPSURL(url, branch)) ? (
        <Link href={getGitRepoHTTPSURL(url, branch)} newTab>
          {linkText}
        </Link>
      ) : (
        <span>{linkText}</span>
      );
    } else {
      return isAllowedLink(getGitRepoHTTPSURL(url)) ? (
        <Link href={getGitRepoHTTPSURL(url)} newTab>
          {linkText}
        </Link>
      ) : (
        <span>{linkText}</span>
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
            {clusters.length === 0 && (
              <MenuItem disabled={true}>Loading...</MenuItem>
            )}
            {clusters?.map((option: GitopsCluster, index: number) => {
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
            {gitRepos?.map((option, index: number) => (
              <MenuItem key={index} value={JSON.stringify(option)}>
                {option.name}&nbsp;&nbsp;
                {optionUrl(option?.url, option?.reference?.branch)}
              </MenuItem>
            ))}
            {helmRepos.length !== 0 && (
              <ListSubheader>HelmRepository</ListSubheader>
            )}
            {helmRepos?.map((option, index: number) => (
              <MenuItem key={index} value={JSON.stringify(option)}>
                {option.name}&nbsp;&nbsp;
                {optionUrl(option?.url)}
              </MenuItem>
            ))}
          </Select>
        </>
      )}
      {formData.source_type === 'GitRepository' || !clusters ? (
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
            name="target_namespace"
            label="TARGET NAMESPACE"
            description="OPTIONAL If omitted all resources must specify a namespace"
            value={formData.clusterAutomations[index].target_namespace}
            onChange={event => handleFormData(event, 'target_namespace')}
          />
          <Input
            className="form-section"
            required={true}
            name="path"
            label="SELECT PATH"
            value={formData.clusterAutomations[index].path}
            onChange={event => handleFormData(event, 'path')}
            description="Path within the git repository to read yaml files"
          />
          {!clusters && (
            <Tooltip
              title="Only the bootstrap GitRepository can be referenced by kustomizations when creating a new cluster"
              placement="bottom-start"
            >
              <span className="input-wrapper">
                <Input
                  className="form-section"
                  type="text"
                  disabled={true}
                  value="flux-system"
                  description="The bootstrap GitRepository object"
                  label="SELECT SOURCE"
                />
              </span>
            </Tooltip>
          )}
        </>
      ) : null}
      {formData.source_type === 'GitRepository' || !clusters ? (
        <Flex align={true}>
          <FormControlLabel
            value="top"
            control={
              <Checkbox
                // Restore default paddingLeft for checkbox that is removed by the global style
                // mui.FormControlLabel does some negative margin to align the checkbox with the label
                style={{ paddingLeft: 9, marginRight: theme.spacing.small }}
                checked={createNamespace}
                onChange={handleCreateNamespace}
                inputProps={{ 'aria-label': 'controlled' }}
                color="primary"
              />
            }
            label="Create target namespace for kustomization"
          />
        </Flex>
      ) : null}
    </FormWrapper>
  );
};

export default AppFields;
