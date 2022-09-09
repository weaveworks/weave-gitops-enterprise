import React, { FC, Dispatch, useState } from 'react';
import styled from 'styled-components';
import _ from 'lodash';
import useProfiles from '../../../../../contexts/Profiles';
import { Input, InputBase, Select } from '../../../../../utils/form';
import {
  createStyles,
  IconButton,
  ListSubheader,
  MenuItem,
  TextField,
} from '@material-ui/core';
import {
  useListSources,
  theme,
  Icon,
  IconType,
} from '@weaveworks/weave-gitops';
import { DEFAULT_FLUX_KUSTOMIZATION_NAMESPACE } from '../../../../../utils/config';
import { getGitRepoHTTPSURL } from '../../../../../utils/formatters';
import { isAllowedLink } from '@weaveworks/weave-gitops';
import { Tooltip } from '../../../../Shared';
import { Autocomplete } from '@material-ui/lab';
import { AddAppFormData } from '../..';
import {
  Source,
  GitRepository,
  HelmRepository,
} from '@weaveworks/weave-gitops/ui/lib/objects';
import { makeStyles } from '@material-ui/styles';
import { theme as weaveTheme } from '@weaveworks/weave-gitops';

function isGitRepository(source: Source): source is GitRepository {
  return source.kind === 'KindGitRepository';
}

function isHelmRepositry(source: Source): source is HelmRepository {
  return source.kind === 'KindHelmRepository';
}

const SourceOption = styled.div`
  display: flex;
  width: 100%;
  white-space: nowrap;
  align-items: center;
  span {
    margin-right: 16px;
  }
  a {
    overflow: hidden;
    text-overflow: ellipsis;
    min-width: 0;
    &:hover {
      text-decoration: underline;
    }
  }
`;

const useStyles = makeStyles(() =>
  createStyles({
    root: {
      marginRight: `${weaveTheme.spacing.medium}`,
    },
    endAdornment: {
      right: 9,
    },
    paper: {
      width: 700,
    },
  }),
);

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
`;

const AppFields: FC<{
  formData: AddAppFormData;
  setFormData: Dispatch<React.SetStateAction<AddAppFormData>>;
  index?: number;
}> = ({ formData, setFormData, index = 0 }) => {
  const { setHelmRepo } = useProfiles();
  const { data, isLoading } = useListSources();
  const automation = formData.clusterAutomations[index];

  const handleSelectCluster = (
    event: React.ChangeEvent<any>,
    clusterName: string | null,
  ) => {
    console.log({ clusterName });
    if (!clusterName) {
      return;
    }
    let currentAutomation = [...formData.clusterAutomations];
    currentAutomation[index] = {
      ...automation,
      clusterName,
    };
    setFormData({
      ...formData,
      clusterAutomations: currentAutomation,
    });
  };

  const clusters = _.uniq(data?.result?.map(s => s.clusterName));

  const clusterSources = _.sortBy(
    data?.result
      .filter(s => s.clusterName === automation.clusterName)
      .filter(s => isGitRepository(s) || isHelmRepositry(s)) || [],
    'name',
    // Still need this even with the type guarded fns?
  ) as (GitRepository | HelmRepository)[];

  const handleSelectSource = (
    event: React.ChangeEvent<any>,
    source: GitRepository | HelmRepository | null,
    reason: string,
  ) => {
    if (!source) {
      return;
    }

    if (event.type == 'blur' && event.target === document.activeElement) {
      return;
    }

    const tagName = (event.target as HTMLElement)?.tagName;
    const clickedLink = tagName === 'A';
    console.log({ tagName, clickedLink, event });

    if (clickedLink) {
      event.stopPropagation();
      return;
    }

    const clusterAutomations = [...formData.clusterAutomations];
    clusterAutomations[index] = {
      ...automation,
      source,
    };

    setFormData({
      ...formData,
      clusterAutomations,
    });

    if (source.kind === 'KindHelmRepository') {
      setHelmRepo({
        name: source.name,
        namespace: source.namespace,
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
    const linkText = branch ? (
      <>
        {url}@<strong>{branch}</strong>
      </>
    ) : (
      url
    );
    return isAllowedLink(getGitRepoHTTPSURL(url, branch)) ? (
      <a
        title="Visit repository"
        style={{
          color: theme.colors.primary,
          fontSize: theme.fontSizes.medium,
        }}
        href={getGitRepoHTTPSURL(url, branch)}
        target="_blank"
        rel="noopener noreferrer"
      >
        {linkText}
      </a>
    ) : (
      <span>{linkText}</span>
    );
  };

  const classes = useStyles();
  const [open, setOpen] = useState(false);
  const handleOpen = (...args: any) => {
    console.log('open', args);
    setOpen(true);
  };
  const handleClose = (event: React.ChangeEvent<{}>) => {
    // Don't close if we've opened a new window
    // https://stackoverflow.com/a/61758013

    if (event.type == 'blur' && event.target === document.activeElement) {
      return;
    }
    const tagName = (event.target as HTMLElement)?.tagName;
    const clickedLink = tagName === 'A';
    if (clickedLink) {
      return;
    }
    setOpen(false);
  };
  console.log(automation.source);
  return (
    <FormWrapper>
      {clusters && (
        <>
          <Input
            name="cluster_name"
            className="form-section"
            label="SELECT CLUSTER"
            description="select target cluster"
          >
            <Autocomplete
              classes={classes}
              loading={isLoading}
              disableClearable
              openOnFocus
              value={automation.clusterName || ''}
              renderInput={({ InputProps, InputLabelProps, ...rest }) => {
                return (
                  <InputBase
                    name="cluster_name"
                    required={true}
                    {...rest}
                    {...InputProps}
                  />
                );
              }}
              options={clusters}
              onChange={handleSelectCluster}
            />
          </Input>
          <Input
            name="source"
            className="form-section"
            label="SELECT SOURCE"
            description="The name and type of source"
          >
            <Autocomplete<GitRepository | HelmRepository>
              classes={classes}
              autoHighlight
              loading={isLoading}
              open={open}
              value={automation.source}
              groupBy={option => option?.type || 'No kind'}
              getOptionSelected={(option, value) =>
                option?.kind === value?.kind &&
                option?.name === value?.name &&
                option?.namespace === value?.namespace
              }
              onChange={handleSelectSource}
              onOpen={handleOpen}
              onClose={handleClose}
              getOptionLabel={option => option?.name || 'hmm'}
              renderOption={option => (
                <SourceOption>
                  <span>{option?.name}</span>
                  {optionUrl(
                    option?.url,
                    (isGitRepository(option) && option?.reference?.branch) ||
                      '',
                  )}
                  {/* <IconButton
                  onClick={(ev: React.MouseEvent) => {
                    ev.stopPropagation();
                    window.open('http://github.com/foot');
                  }}
                >
                  <Icon type={IconType.ExternalTab} size="base" />
                </IconButton> */}
                </SourceOption>
              )}
              options={clusterSources}
              renderInput={({ InputProps, InputLabelProps, ...rest }) => {
                return (
                  <InputBase
                    name="source"
                    required={true}
                    {...rest}
                    {...InputProps}
                  />
                );
              }}
            />
          </Input>
          {/* {[...gitRepos, ...helmRepos].length === 0 && (
              <MenuItem disabled={true}>
                No repository available, please select another cluster.
              </MenuItem>
            )}
            {gitRepos.length !== 0 && (
              <ListSubheader>GitRepository</ListSubheader>
            )}
            {gitRepos?.map((option: GitRepository, index: number) => (
              <MenuItem key={index} value={JSON.stringify(option)}>
                {option.name}&nbsp;&nbsp;
                {optionUrl(option?.url, option?.reference?.branch)}
              </MenuItem>
            ))}
            {helmRepos.length !== 0 && (
              <ListSubheader>HelmRepository</ListSubheader>
            )}
            {helmRepos?.map((option: HelmRepository, index: number) => (
              <MenuItem key={index} value={JSON.stringify(option)}>
                {option.name}&nbsp;&nbsp;
                {optionUrl(option?.url)}
              </MenuItem>
            ))} */}
        </>
      )}
      {automation.source?.kind === 'KindGitRepository' || !clusters ? (
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
    </FormWrapper>
  );
};

export default AppFields;
