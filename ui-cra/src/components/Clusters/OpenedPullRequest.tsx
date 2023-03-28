import React, { useContext, useMemo } from 'react';
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
import { GitAuth } from '../../contexts/GitAuth';
import GitUrlParse from 'git-url-parse';
import { openLinkHandler } from '../../utils/link-checker';
import useConfig from '../../hooks/config';
// import urijs
import URI from 'urijs';
import { GetConfigResponse } from '../../cluster-services/cluster_services.pb';
import { getInitialGitRepo } from '../Templates/Form/utils';

type Props = {
  gitRepos: GitRepository[];
};

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

function getDomain(url: string) {
  const uri = URI(url);
  return uri.domain();
}

function getHttpsUrl(url: string) {
  // FIXME: does not work for azure devops
  const uri = URI(url);
  return uri.protocol('https').toString();
}

function getPullRequestUrl(gitRepo: GitRepository, config: GetConfigResponse) {
  // FIXME: check the annoation as well
  console.log({ gitRepo });

  const domain = getDomain(gitRepo.obj.spec.url);
  console.log({ gitRepo, domain });
  const provider = config?.gitHostTypes?.[domain] || 'github';
  if (provider === 'github') {
    return getHttpsUrl(gitRepo.obj.spec.url) + '/pulls';
  }
  if (provider === 'gitlab') {
    return getHttpsUrl(gitRepo.obj.spec.url) + '/-/merge_requests';
  }

  // FIXME: this is not correct
  if (provider === 'bitbucket') {
    return getHttpsUrl(gitRepo.obj.spec.url) + '/pull-requests';
  }

  // FIXME: this is not correct
  if (provider === 'azuredevops') {
    return getHttpsUrl(gitRepo.obj.spec.url) + '/pullrequests';
  }
}

export default function OpenedPullRequest({ gitRepos }: Props) {
  const [open, setOpen] = React.useState(false);
  const anchorRef = React.useRef<HTMLDivElement>(null);
  const [selectedIndex, setSelectedIndex] = React.useState(-1);
  const { gitAuthClient } = useContext(GitAuth);
  const [OpenPrUrl, setOpenPrUrl] = React.useState('');
  const [OpenPrButtonDisabled, setOpenPrButtonDisabled] = React.useState(false);

  const options = useMemo(
    () =>
      gitRepos.map(
        repo =>
          repo?.obj?.metadata?.annotations?.['weave.works/repo-https-url'] ||
          repo.obj.spec.url,
      ),
    [gitRepos],
  );

  const Classes = useStyles();

  const { data: config, isLoading } = useConfig();
  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (!config) {
    return <div>Config not found</div>;
  }

  console.log({ gitRepos });
  const initialUrl = getInitialGitRepo('', gitRepos) as GitRepository;
  const initialPullRequestUrl = getPullRequestUrl(initialUrl, config) || '';

  const handleMenuItemClick = (
    event: React.MouseEvent<HTMLLIElement, MouseEvent>,
    index: number,
  ) => {
    setSelectedIndex(index);
    setOpen(false);
    const repoUrl = options[index];
    setOpenPrButtonDisabled(true);
    gitAuthClient.ParseRepoURL({ url: repoUrl }).then(res => {
      setOpenPrButtonDisabled(false);
      const { protocol, href } = GitUrlParse(repoUrl);
      let parsedUrl = '';
      if (protocol === 'ssh') {
        parsedUrl = href.replace('ssh://git@', 'https://');
      }
      const provider = res.provider || '';
      if (provider === 'GitHub') {
        setOpenPrUrl(`${parsedUrl}/pulls`);
      }
      if (provider === 'GitLab') {
        setOpenPrUrl(`${parsedUrl}/-/merge_requests`);
      }
    });
  };

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
          onClick={openLinkHandler(initialPullRequestUrl)}
          disabled={OpenPrButtonDisabled || !options.length}
        >
          {selectedIndex === -1 ? (
            'SELECT GIT REPOSITORY'
          ) : (
            <>
              <Icon
                className={Classes.externalLink}
                type={IconType.ExternalTab}
                size="base"
              />
              GO TO OPEN PULL REQUESTS AT {options[selectedIndex]}
            </>
          )}
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
                      selected={index === selectedIndex}
                      // FIXME: change to openLinkHandler
                      onClick={event => handleMenuItemClick(event, index)}
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
