import React, { useMemo } from 'react';
import { Button, GitRepository } from '@weaveworks/weave-gitops';
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
} from '../Templates/Form/utils';
import { useGitRepos } from '../../hooks/gitrepos';

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

function getPullRequestUrl(gitRepo: GitRepository, config: GetConfigResponse) {
  const provider = getProvider(gitRepo, config);

  const baseUrl = getRepositoryUrl(gitRepo);
  if (provider === 'gitlab') {
    return baseUrl + '/-/merge_requests';
  }

  // FIXME: this is not correct
  if (provider === 'bitbucket') {
    return baseUrl + '/pull-requests';
  }

  // FIXME: this is not correct
  if (provider === 'azuredevops') {
    return baseUrl + '/pullrequests';
  }

  // github is the default
  return baseUrl + '/pulls';
}

export default function OpenedPullRequest() {
  const [open, setOpen] = React.useState(false);
  const anchorRef = React.useRef<HTMLDivElement>(null);

  const { gitRepos } = useGitRepos();

  const Classes = useStyles();

  const { data: config, isLoading } = useConfig();

  const options = useMemo(
    () =>
      !config
        ? ([] as string[])
        : gitRepos.map(repo => getPullRequestUrl(repo, config)),
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

  const defaultRepo = getDefaultGitRepo(gitRepos, config);

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
          View open pull requests
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
                    <MenuItem
                      key={option}
                      // FIXME: change to openLinkHandler
                      onClick={openLinkHandler(option)}
                    >
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
