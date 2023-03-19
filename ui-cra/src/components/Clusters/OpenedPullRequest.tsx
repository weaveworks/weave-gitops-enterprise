import React, { useContext } from 'react';
import { Button, Icon, IconType } from '@weaveworks/weave-gitops';
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

type Props = {
  options: string[];
};

export default function OpenedPullRequest({ options }: Props) {
  const [open, setOpen] = React.useState(false);
  const anchorRef = React.useRef<HTMLDivElement>(null);
  const [selectedIndex, setSelectedIndex] = React.useState(-1);
  const { gitAuthClient } = useContext(GitAuth);
  const [OpenPrUrl, setOpenPrUrl] = React.useState('');
  const [OpenPrButtonDisabled, setOpenPrButtonDisabled] = React.useState(false);

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
      const { resource, full_name, protocol } = GitUrlParse(repoUrl);
      const provider = res.provider || '';
      if (provider === 'GitHub') {
        setOpenPrUrl(`${protocol}://${resource}/${full_name}/pulls`);
      }
      if (provider === 'GitLab') {
        setOpenPrUrl(`${protocol}://${resource}/${full_name}/-/merge_requests`);
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
  const Classes = useStyles();
  return (
    <>
      <ButtonGroup variant="outlined" ref={anchorRef} aria-label="split button">
        <Button
          className={Classes.optionsButton}
          color="primary"
          onClick={openLinkHandler(OpenPrUrl)}
          disabled={OpenPrButtonDisabled || options.length === 0}
        >
          {selectedIndex === -1 ? (
            'SELECT GIT REPOSIOTORY'
          ) : (
            <>
              <Icon
                className={Classes.externalLink}
                type={IconType.ExternalTab}
                size="base"
              />
              GO TO OPEN PULL REQUESTS ON {options[selectedIndex]}
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
