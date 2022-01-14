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
import { theme as weaveTheme, Button, Icon } from '@weaveworks/weave-gitops';

const medium = weaveTheme.spacing.medium;
const xs = weaveTheme.spacing.xs;

const useStyles = makeStyles(() => ({
  textarea: {
    width: '100%',
    padding: xs,
    border: `1px solid ${weaveTheme.colors.neutral10}`,
  },
}));

const ListItemWrapper = styled.div`
  & .profile-name {
    margin-right: ${medium};
  }
  & .profile-version,
  .profile-layer {
    display: flex;
    align-items: center;
    margin-right: ${medium};
    span {
      margin-right: ${xs};
    }
  }
  & .profile-version {
    width: 150px;
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
        <ListItem data-profile-name={profile.name}>
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
          <div className="profile-layer">
            <span>Layer</span>
            <span>Example layer</span>
            {/* {profile.layer} */}
          </div>
          <Button variant="text" onClick={handleYamlPreview}>
            Values.yaml
          </Button>
        </ListItem>
      </ListItemWrapper>

      <Dialog
        open={openYamlPreview}
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
          <Button
            id="edit-yaml"
            startIcon={<Icon type="SaveAlt" size="base" />}
            onClick={handleUpdateProfiles}
            disabled={profile.required}
          >
            SAVE CHANGES
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
};

export default ProfilesListItem;
