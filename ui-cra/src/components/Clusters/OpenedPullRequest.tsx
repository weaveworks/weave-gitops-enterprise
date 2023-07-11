import React, { useMemo } from 'react';
import {
  Button,
  GitRepository,
  Icon,
  IconType,
} from '@weaveworks/weave-gitops';
import ButtonGroup from '@material-ui/core/ButtonGroup';
import ArrowDropDownIcon from '@material-ui/icons/ArrowDropDown';
import ClickAwayListener from '@material-ui/core/ClickAwayListener';
import Grow from '@material-ui/core/Grow';
import Paper from '@material-ui/core/Paper';
import Popper from '@material-ui/core/Popper';
import MenuItem from '@material-ui/core/MenuItem';
import MenuList from '@material-ui/core/MenuList';
import { createStyles, makeStyles } from '@material-ui/core';
import { openLinkHandler } from '../../utils/link-checker';
import useConfig from '../../hooks/config';
import { GetConfigResponse } from '../../cluster-services/cluster_services.pb';
import {
  getDefaultGitRepo,
  getProvider,
  getRepositoryUrl,
  bitbucketReposToHttpsUrl,
} from '../Templates/Form/utils';
import { useGitRepos } from '../../hooks/gitrepos';
import { useListConfigContext } from '../../contexts/ListConfig';
import _ from 'lodash';

const useStyles = makeStyles(() =>
  createStyles({
    optionsButton: {
      marginRight: '0px',
    },
    externalLink: {
      marginRight: '5px',
    },
  }),
);

export function getPullRequestUrl(
  gitRepo: GitRepository,
  config: GetConfigResponse,
) {
  console.log('gitRepo', gitRepo);
  const provider = getProvider(gitRepo, config);

  const repoUrl = getRepositoryUrl(gitRepo);

  // remove any trailing .git
  const baseUrl = repoUrl.replace(/\.git$/, '');

  if (provider === 'gitlab') {
    return baseUrl + '/-/merge_requests';
  }

  if (provider === 'bitbucket-server') {
    const url = bitbucketReposToHttpsUrl(baseUrl);

    return url + '/pull-requests';
  }

  if (provider === 'azure-devops') {
    return baseUrl + '/pullrequests?_a=active';
  }

  // github is the default
  return baseUrl + '/pulls';
}

export default function OpenedPullRequest() {
  const [open, setOpen] = React.useState(false);
  const configResponse = useListConfigContext();
  const mgCluster = configResponse?.data?.managementClusterName;

  const anchorRef = React.useRef<HTMLDivElement>(null);

  const { gitRepos } = useGitRepos();

  const Classes = useStyles();

  const { data: config, isLoading } = useConfig();

  const options = useMemo(
    () =>
      !config
        ? ([] as string[])
        : _.uniq(gitRepos.map(repo => getPullRequestUrl(repo, config))),
    [gitRepos, config],
  );

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (!config) {
    return <div>Config not found</div>;
  }

  if (!gitRepos || gitRepos.length === 0) {
    return <div>Git Repos not found</div>;
  }

  const defaultRepo = getDefaultGitRepo(gitRepos, mgCluster);

  const handleToggle = () => {
    setOpen(prevOpen => !prevOpen);
  };

  const handleClose = (event: React.MouseEvent<Document, MouseEvent>) => {
    if (
      anchorRef.current &&
      anchorRef.current.contains(event.target as HTMLElement)
    ) {
      return;
    }

    setOpen(false);
  };

  return (
    <>
      <ButtonGroup variant="outlined" ref={anchorRef} aria-label="split button">
        <Button
          className={Classes.optionsButton}
          color="primary"
          onClick={openLinkHandler(
            getPullRequestUrl(defaultRepo, config) || '',
          )}
          disabled={!options.length}
        >
          <>
            <Icon
              className={Classes.externalLink}
              type={IconType.ExternalTab}
              size="base"
            />
            VIEW OPEN PULL REQUESTS
          </>
        </Button>
        <Button
          size="small"
          aria-controls={open ? 'split-button-menu' : undefined}
          aria-expanded={open ? 'true' : undefined}
          aria-haspopup="menu"
          onClick={handleToggle}
          color="primary"
          disabled={options.length === 0}
        >
          <ArrowDropDownIcon />
        </Button>
      </ButtonGroup>
      <Popper
        open={open}
        anchorEl={anchorRef.current}
        role={undefined}
        transition
      >
        {({ TransitionProps, placement }) => (
          <Grow
            {...TransitionProps}
            style={{
              transformOrigin:
                placement === 'bottom' ? 'center top' : 'center bottom',
            }}
          >
            <Paper>
              <ClickAwayListener onClickAway={handleClose}>
                <MenuList id="split-button-menu">
                  {options.map((option, index) => (
                    <MenuItem key={option} onClick={openLinkHandler(option)}>
                      {option}
                    </MenuItem>
                  ))}
                </MenuList>
              </ClickAwayListener>
            </Paper>
          </Grow>
        )}
      </Popper>
    </>
  );
}
