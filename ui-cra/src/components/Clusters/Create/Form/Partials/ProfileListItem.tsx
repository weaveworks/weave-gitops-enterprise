import React, {
  ChangeEvent,
  FC,
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
  FormControl,
  Select,
  MenuItem,
} from '@material-ui/core';
import Typography from '@material-ui/core/Typography';
import { CloseIconButton } from '../../../../../assets/img/close-icon-button';
import { OnClickAction } from '../../../../Action';
import weaveTheme from 'weaveworks-ui-components/lib/theme';
import Button from '@material-ui/core/Button';
import { GitOpsBlue } from '../../../../../muiTheme';

const medium = weaveTheme.spacing.medium;
const xs = weaveTheme.spacing.xs;

const useStyles = makeStyles(() => ({
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
  & .profile-name {
    margin-right: ${medium};
  }
  & .profile-version {
    display: flex;
    align-items: center;
    margin-right: ${medium};
    width: 150px;
    span {
      margin-right: ${xs};
    }
  }
`;

const ProfilesListItem: FC<{
  profile: UpdatedProfile;
  updateProfile: (profile: UpdatedProfile) => void;
}> = ({ profile, updateProfile }) => {
  const classes = useStyles();
  const [version, setVersion] = useState<string>('');
  const [yaml, setYaml] = useState<string>('');
  const [openYamlPreview, setOpenYamlPreview] = useState<boolean>(false);

  const profileVersions = (profile: UpdatedProfile) => [
    ...profile.values.map((value, index) => {
      const { version } = value;
      return (
        <MenuItem key={index} value={version}>
          {version}
        </MenuItem>
      );
    }),
  ];

  const handleSelectVersion = useCallback(
    (event: ChangeEvent<{ name?: string | undefined; value: unknown }>) => {
      const value = event.target.value as string;
      setVersion(value);

      profile.values.forEach(item =>
        item.selected === true ? (item.selected = false) : null,
      );

      profile.values.forEach(item => {
        if (item.version === value) {
          item.selected = true;
          setYaml(item.yaml as string);
          return;
        }
      });

      updateProfile(profile);
    },
    [profile, updateProfile],
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
    profile.values.forEach(item => {
      if (item.version === version) {
        item.yaml = yaml;
      }
    });

    updateProfile(profile);

    setOpenYamlPreview(false);
  }, [profile, updateProfile, version, yaml]);

  useEffect(() => {
    setVersion(profile.values[0].version as string);
    setYaml(profile.values[0].version as string);
    profile.values[0].selected = true;
  }, [profile]);

  return (
    <>
      <ListItemWrapper>
        <ListItem>
          <ListItemText className="profile-name">{profile.name}</ListItemText>
          <div className="profile-version">
            <span>Version</span>
            <FormControl>
              <Select
                disabled={profile.required}
                value={version}
                onChange={handleSelectVersion}
                autoWidth
                label="Versions"
              >
                {profileVersions(profile)}
              </Select>
            </FormControl>
          </div>
          <Button className={classes.downloadBtn} onClick={handleYamlPreview}>
            Values.yaml
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
            disabled={profile.required}
          />
        </DialogActions>
      </Dialog>
    </>
  );
};

export default ProfilesListItem;
