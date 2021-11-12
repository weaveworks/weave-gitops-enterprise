import React, {
  ChangeEvent,
  FC,
  FormEvent,
  useCallback,
  useEffect,
  useState,
} from 'react';
import styled from 'styled-components';
import ListItemText from '@material-ui/core/ListItemText';
import { makeStyles } from '@material-ui/core/styles';
import { UpdatedProfile } from '../../../../../types/custom';
import ListItem from '@material-ui/core/ListItem';
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogActions,
  TextareaAutosize,
} from '@material-ui/core';
import Typography from '@material-ui/core/Typography';
import { CloseIconButton } from '../../../../../assets/img/close-icon-button';
import { OnClickAction } from '../../../../Action';
import weaveTheme from 'weaveworks-ui-components/lib/theme';
import Button from '@material-ui/core/Button';
import { GitOpsBlue } from '../../../../../muiTheme';
import { Dropdown } from 'weaveworks-ui-components';

const medium = weaveTheme.spacing.medium;
const xs = weaveTheme.spacing.xs;

const useStyles = makeStyles(theme => ({
  dialog: {
    backgroundColor: weaveTheme.colors.gray50,
  },
  textarea: {
    width: '100%',
    padding: xs,
    border: '1px solid #E5E5E5',
  },
  downloadBtn: {
    color: GitOpsBlue,
    padding: '0px',
  },
}));

const ListItemWrapper = styled.div`
  display: flex;
  align-items: center;
  & .profile-name {
    margin-right: ${medium};
  }
  & .profile-version {
    display: flex;
    align-items: center;
    margin-right: ${xs};
    width: 150px;
    span {
      margin-right: ${xs};
    }
  }
  & .dropdown-toggle {
    border: 1px solid #e5e5e5;
  }
`;

const ProfilesListItem: FC<{
  profile: UpdatedProfile;
  updateProfile: (profile: UpdatedProfile) => void;
}> = ({ profile, updateProfile }) => {
  const classes = useStyles();
  const [currentProfileName, setCurrentProfileName] = useState<string>('');
  const [version, setVersion] = useState<string>('');
  const [yaml, setYaml] = useState<string>('');
  const [openYamlPreview, setOpenYamlPreview] = useState<boolean>(false);
  const [enrichedProfiles, setEnrichedProfiles] = useState<UpdatedProfile[]>(
    [],
  );

  const profileVersions = (profile: UpdatedProfile) => [
    ...profile.values.map(value => {
      const { version } = value;
      return {
        label: version as string,
        value: version as string,
      };
    }),
    // { label: 'Select', value: '' },
  ];

  const handleSelectVersion = useCallback(
    (event: FormEvent<HTMLInputElement>, value: string) => {
      setVersion(value);

      // !!! should trigger the :
      //   - update of the relevant yaml
      //   - update of enrichedProfiles here!!!

      //   const currentProfileIndex = enrichedProfiles.findIndex(
      //     p => p.name === profile.name,
      //   );

      //   const [currentValue] = enrichedProfiles[
      //     currentProfileIndex
      //   ].values.filter(item => item.version === value);

      //   currentValue.selected = true;
    },
    [],
  );

  const handleYamlPreview = () => {
    const currentProfile = profile.values.find(
      value => value.version === version,
    );
    setYaml(currentProfile?.yaml as string);
    setOpenYamlPreview(true);
  };

  const handleChangeYaml = (event: ChangeEvent<HTMLTextAreaElement>) =>
    setYaml(event.target.value);

  const handleUpdateProfiles = useCallback(() => {
    // !!! should trigger the :
    //   - update of enrichedProfiles here!!!

    setOpenYamlPreview(false);
  }, []);

  const selectedVersionYaml = (profile: UpdatedProfile) =>
    profile.values.find(value => value.selected === true);

  useEffect(() => {
    if (profile.values.length === 1) {
      setVersion(profile.values[0].version as string);
      setYaml(profile.values[0].version as string);
    } else {
      setVersion('');
      setYaml('');
    }
  }, [profile]);

  return (
    <>
      <ListItemWrapper>
        <ListItem className="">
          <ListItemText className="profile-name">{profile.name}</ListItemText>
          <div className="profile-version">
            <span>Version</span>
            <Dropdown
              value={version}
              items={profileVersions(profile)}
              onChange={(event, value) => handleSelectVersion(event, value)}
            />
          </div>
          <Button
            className={classes.downloadBtn}
            disabled={version === ''}
            onClick={handleYamlPreview}
          >
            values.yaml
          </Button>
        </ListItem>
      </ListItemWrapper>

      <Dialog
        open={openYamlPreview}
        className={classes.dialog}
        maxWidth="md"
        fullWidth
        scroll="paper"
        onClose={() => setOpenYamlPreview(false)}
      >
        <DialogTitle disableTypography>
          <Typography variant="h5">{profile.name}</Typography>
          <CloseIconButton onClick={() => setOpenYamlPreview(false)} />
        </DialogTitle>
        <DialogContent>
          <TextareaAutosize
            className={classes.textarea}
            defaultValue={yaml}
            onChange={event => handleChangeYaml(event)}
          />
        </DialogContent>
        <DialogActions>
          <OnClickAction
            id="edit-yaml"
            onClick={handleUpdateProfiles}
            text="Save changes"
          />
        </DialogActions>
      </Dialog>
    </>
  );
};

export default ProfilesListItem;
