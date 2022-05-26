import React, {
  ChangeEvent,
  FC,
  useCallback,
  useEffect,
  useState,
} from 'react';
import styled from 'styled-components';
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
import {
  theme as weaveTheme,
  Button,
  Icon,
  IconType,
} from '@weaveworks/weave-gitops';

const base = weaveTheme.spacing.base;
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
  & .profile-version,
  .profile-layer {
    display: flex;
    align-items: center;
    margin-left: ${base};
    span {
      margin-right: ${xs};
    }
  }
  ,
  & .profile-name,
  .profile-layer {
    min-width: 120px;
  }
  & .profile-version {
    .MuiSelect-root {
      min-width: 75px;
    }
  }
`;

const ProfilesListItem: FC<{
  profile: UpdatedProfile;
  updateProfile: (profile: UpdatedProfile) => void;
}> = ({ profile, updateProfile }) => {
  console.log(profile);
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
    if (profile.values.filter(value => value.selected === true).length > 0) {
      setVersion(
        profile.values.filter(value => value.selected === true)[0].version,
      );
    } else {
      setVersion(profile.values[0].version as string);
      setYaml(profile.values[0].version as string);
      profile.values[0].selected = true;
    }
  }, [profile]);

  return (
    <>
      <ListItemWrapper>
        <ListItem data-profile-name={profile.name}>
          <div className="profile-name">{profile.name}</div>
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
          <Button
            style={{ marginLeft: medium }}
            variant="text"
            onClick={handleYamlPreview}
          >
            Values.yaml
          </Button>
          {profile.layer ? (
            <div className="profile-layer">
              <span>Layer</span>
              <span>{profile.layer}</span>
            </div>
          ) : null}
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
            startIcon={<Icon type={IconType.SaveAltIcon} size="base" />}
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
