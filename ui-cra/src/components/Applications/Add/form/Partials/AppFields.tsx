import React, { FC, Dispatch, useEffect, useCallback } from 'react';
import { useHistory, useLocation } from 'react-router-dom';
import styled from 'styled-components';
import _ from 'lodash';
import { Input, Select } from '../../../../../utils/form';
import {
  ListSubheader,
  MenuItem,
  Checkbox,
  FormControlLabel,
} from '@material-ui/core';
import { useListSources, theme, Flex, Kind } from '@weaveworks/weave-gitops';
import { DEFAULT_FLUX_KUSTOMIZATION_NAMESPACE } from '../../../../../utils/config';
import {
  GitRepository,
  HelmRepository,
} from '@weaveworks/weave-gitops/ui/lib/objects';
import { Tooltip } from '../../../../Shared';
import { GitopsCluster } from '../../../../../cluster-services/cluster_services.pb';
import { useClustersWithSources } from '../../../utils';
import { GitopsFormData } from '../../../../Templates/Form/utils';

const AppFieldsWrapper = styled.div`
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
  formData: GitopsFormData;
  setFormData: Dispatch<React.SetStateAction<GitopsFormData>>;
  index?: number;
  onPRPreview?: () => void;
  previewLoading?: boolean;
  allowSelectCluster: boolean;
  context?: string;
  clusterName?: string;
  formError?: string;
}> = ({
  formData,
  setFormData,
  index = 0,
  allowSelectCluster,
  clusterName,
  formError,
}) => {
  const { data } = useListSources();
  const { source } = formData;
  const app = formData.clusterAutomations[index];
  const history = useHistory();
  const location = useLocation();

  let clusters: GitopsCluster[] | undefined =
    useClustersWithSources(allowSelectCluster);

  const updateCluster = useCallback(
    (cluster: GitopsCluster) => {
      setFormData(formData => {
        const params = new URLSearchParams(`clusterName=${cluster.name}`);
        history.replace({
          pathname: location.pathname,
          search: params.toString(),
        });
        const newAutomations = [...formData.clusterAutomations];
        newAutomations[index] = {
          ...newAutomations[index],
          cluster_name: cluster.name!,
          cluster_namespace: cluster.namespace!,
          cluster_isControlPlane: Boolean(cluster.controlPlane),
          cluster: JSON.stringify(cluster),
        };
        return {
          ...formData,
          clusterAutomations: newAutomations,
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
    const clusterName = app.cluster_namespace
      ? `${app.cluster_namespace}/${app.cluster_name}`
      : `${app.cluster_name}`;

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

    const selectedSource = {
      name: obj?.metadata?.name,
      namespace: obj?.metadata?.namespace,
      type: obj?.kind,
      url: obj?.spec.url,
      branch: obj?.kind === 'GitRepository' ? obj?.spec.ref.branch : '',
      data: value,
    };

    const newAutomations = [...formData.clusterAutomations];
    newAutomations[index] = {
      ...newAutomations[index],
    };

    setFormData({
      ...formData,
      source: selectedSource,
      clusterAutomations: newAutomations,
    });
  };

  const handleFormData = (
    event:
      | React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>
      | React.ChangeEvent<{ name?: string; value: unknown }>,
    fieldName?: string,
  ) => {
    const { value } = event?.target;

    let newAutomations = [...formData.clusterAutomations];
    newAutomations[index] = {
      ...newAutomations[index],
      [fieldName as string]: value,
    };

    // Special case, don't allow the user to try and re-create the flux-system namespace
    // it will be on every cluster already
    if (fieldName === 'target_namespace' && value === 'flux-system') {
      newAutomations[index] = {
        ...newAutomations[index],
        createNamespace: false,
      };
    }

    setFormData({
      ...formData,
      clusterAutomations: newAutomations,
    });
  };

  const handleCreateNamespace = (
    event: React.ChangeEvent<HTMLInputElement>,
  ) => {
    const newAutomations = [...formData.clusterAutomations];
    newAutomations[index] = {
      ...newAutomations[index],
      createNamespace: event.target.checked,
    };

    setFormData({
      ...formData,
      clusterAutomations: newAutomations,
    });
  };

  return (
    <AppFieldsWrapper>
      {!!clusters && (
        <>
          <Select
            className="form-section"
            name="cluster_name"
            required={true}
            label="SELECT CLUSTER"
            value={app.cluster || ''}
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
            value={source.data || ''}
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
                {option.name}
              </MenuItem>
            ))}
            {helmRepos.length !== 0 && (
              <ListSubheader>HelmRepository</ListSubheader>
            )}
            {helmRepos?.map((option, index: number) => (
              <MenuItem key={index} value={JSON.stringify(option)}>
                {option.name}
              </MenuItem>
            ))}
          </Select>
        </>
      )}
      {formData.source.type === 'GitRepository' || !clusters ? (
        <>
          <Input
            className="form-section"
            required={true}
            name="name"
            label="KUSTOMIZATION NAME"
            value={app.name}
            onChange={event => handleFormData(event, 'name')}
            error={formError === 'name' && !app.name}
          />
          <Input
            className="form-section"
            name="namespace"
            label="KUSTOMIZATION NAMESPACE"
            placeholder={DEFAULT_FLUX_KUSTOMIZATION_NAMESPACE}
            value={app.namespace}
            onChange={event => handleFormData(event, 'namespace')}
            error={formError === 'namespace' && !app.namespace}
          />
          <Input
            className="form-section"
            name="target_namespace"
            label="TARGET NAMESPACE"
            description="OPTIONAL If omitted all resources must specify a namespace"
            value={app.target_namespace}
            onChange={event => handleFormData(event, 'target_namespace')}
            error={formError === 'target_namespace' && !app.target_namespace}
          />
          <Input
            className="form-section"
            required={true}
            name="path"
            label="SELECT PATH"
            value={app.path}
            onChange={event => handleFormData(event, 'path')}
            description="Path within the git repository to read yaml files"
            error={formError === 'path' && !app.path}
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
      {formData.source.type === 'GitRepository' || !clusters ? (
        <Tooltip
          title={'flux-system will already exist in the target cluster'}
          placement="top-start"
          disabled={app.target_namespace !== 'flux-system'}
        >
          <div>
            <Flex align={true}>
              <FormControlLabel
                value="top"
                control={
                  <Checkbox
                    // Restore default paddingLeft for checkbox that is removed by the global style
                    // mui.FormControlLabel does some negative margin to align the checkbox with the label
                    style={{ paddingLeft: 9, marginRight: theme.spacing.small }}
                    checked={app.createNamespace}
                    onChange={handleCreateNamespace}
                    inputProps={{ 'aria-label': 'controlled' }}
                    color="primary"
                    disabled={app.target_namespace === 'flux-system'}
                  />
                }
                label="Create target namespace for kustomization"
              />
            </Flex>
          </div>
        </Tooltip>
      ) : null}
    </AppFieldsWrapper>
  );
};

export default AppFields;
