import {
  Checkbox,
  FormControlLabel,
  ListSubheader,
  MenuItem,
} from '@material-ui/core';
import { Flex, Kind, theme, useListSources } from '@weaveworks/weave-gitops';
import {
  GitRepository,
  HelmRepository,
} from '@weaveworks/weave-gitops/ui/lib/objects';
import _ from 'lodash';
import React, { Dispatch, FC, useCallback, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { GitopsCluster } from '../../../../../cluster-services/cluster_services.pb';
import { DEFAULT_FLUX_KUSTOMIZATION_NAMESPACE } from '../../../../../utils/config';
import { Input, Select } from '../../../../../utils/form';
import { Tooltip } from '../../../../Shared';
import { useClustersWithSources } from '../../../utils';

const AppFieldsWrapper = styled.div`
  .form-section {
    width: 50%;
  }
  .input-wrapper {
    padding-bottom: ${({ theme }) => theme.spacing.medium};
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
  setHelmRepo?: Dispatch<React.SetStateAction<any>>;
  formError?: string;
}> = ({
  formData,
  setFormData,
  index = 0,
  allowSelectCluster,
  clusterName,
  setHelmRepo,
  formError,
}) => {
  const { data } = useListSources();
  const automation = formData.clusterAutomations[index];
  const { cluster, source, name, namespace, target_namespace, path } =
    formData.clusterAutomations[index];
  const { createNamespace } = automation;
  // const history = useHistory();
  const navigate = useNavigate();
  const location = useLocation();

  let clusters: GitopsCluster[] | undefined =
    useClustersWithSources(allowSelectCluster);

  const updateCluster = useCallback(
    (cluster: GitopsCluster) => {
      setFormData((formData: any) => {
        const params = new URLSearchParams(`clusterName=${cluster.name}`);
        navigate({
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
      source_type: obj?.kind,
      source: value,
    };

    setFormData({
      ...formData,
      source_name: obj?.metadata?.name,
      source_namespace: obj?.metadata?.namespace,
      source_type: obj?.kind,
      source_url: obj?.spec.url,
      source_branch: obj?.kind === 'GitRepository' ? obj?.spec.ref.branch : '',
      source: value,
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
    if (fieldName === 'target_namespace' && value === 'flux-system') {
      currentAutomation[index] = {
        ...currentAutomation[index],
        createNamespace: false,
      };
    }

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

  return (
    <AppFieldsWrapper>
      {!!clusters && (
        <>
          <Select
            className="form-section"
            name="cluster_name"
            required={true}
            label="SELECT CLUSTER"
            value={cluster || ''}
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
            value={source || ''}
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
      {formData.source_type === 'GitRepository' || !clusters ? (
        <>
          <Input
            className="form-section"
            required={true}
            name="name"
            label="KUSTOMIZATION NAME"
            value={name}
            onChange={event => handleFormData(event, 'name')}
            error={formError === 'name' && !name}
          />
          <Input
            className="form-section"
            name="namespace"
            label="KUSTOMIZATION NAMESPACE"
            placeholder={DEFAULT_FLUX_KUSTOMIZATION_NAMESPACE}
            value={namespace}
            onChange={event => handleFormData(event, 'namespace')}
            error={formError === 'namespace' && !namespace}
          />
          <Input
            className="form-section"
            name="target_namespace"
            label="TARGET NAMESPACE"
            description="OPTIONAL If omitted all resources must specify a namespace"
            value={target_namespace}
            onChange={event => handleFormData(event, 'target_namespace')}
            error={formError === 'target_namespace' && !target_namespace}
          />
          <Input
            className="form-section"
            required={true}
            name="path"
            label="SELECT PATH"
            value={path}
            onChange={event => handleFormData(event, 'path')}
            description="Path within the git repository to read yaml files"
            error={formError === 'path' && !path}
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
        <Tooltip
          title={'flux-system will already exist in the target cluster'}
          placement="top-start"
          disabled={target_namespace !== 'flux-system'}
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
                    checked={createNamespace}
                    onChange={handleCreateNamespace}
                    inputProps={{ 'aria-label': 'controlled' }}
                    color="primary"
                    disabled={target_namespace === 'flux-system'}
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
