import React, {
  ChangeEvent,
  FC,
  useCallback,
  useEffect,
  useState,
} from 'react';
import styled from 'styled-components';
import { makeStyles } from '@material-ui/core/styles';
import {
  ListProfileValuesResponse,
  UpdatedProfile,
} from '../../../../../types/custom';
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
  Input,
} from '@material-ui/core';
import Typography from '@material-ui/core/Typography';
import { CloseIconButton } from '../../../../../assets/img/close-icon-button';
import {
  theme as weaveTheme,
  Button,
  Icon,
  IconType,
} from '@weaveworks/weave-gitops';
import useProfiles from './../../../../../contexts/Profiles';
import { Loader } from '../../../../Loader';

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
  .profile-layer,
  .profile-namespace {
    display: flex;
    align-items: center;
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
  selectedProfiles: UpdatedProfile[];
  setSelectedProfiles: any;
}> = ({ profile, selectedProfiles, setSelectedProfiles }) => {
  const classes = useStyles();
  const [version, setVersion] = useState<string>('');
  const [yaml, setYaml] = useState<string>('');
  const [openYamlPreview, setOpenYamlPreview] = useState<boolean>(false);
  const [namespace, setNamespace] = useState<string>();
  const [isNamespaceValid, setNamespaceValidation] = useState<boolean>(true);
  const [loadingYaml, setLoadingYaml] = useState<boolean>(false);
  const { getProfileYaml } = useProfiles();

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

  const handleUpdateProfile = useCallback(
    profile => {
      const currentProfileIndex = selectedProfiles.findIndex(
        p => p.name === profile.name,
      );
      selectedProfiles[currentProfileIndex] = profile;
      setSelectedProfiles(selectedProfiles);
    },
    [setSelectedProfiles, selectedProfiles],
  );

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

      handleUpdateProfile(profile);
    },
    [profile, handleUpdateProfile],
  );

  const handleYamlPreview = () => {
    setOpenYamlPreview(true);
    if (yaml === '') {
      setLoadingYaml(true);
      getProfileYaml(profile?.name, version)
        .then((res: ListProfileValuesResponse) => setYaml(res.message))
        .finally(() => setLoadingYaml(false));
    }
  };
  const handleChangeNamespace = (event: ChangeEvent<HTMLInputElement>) => {
    const { value } = event.target;
    const pattern = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/;
    if (pattern.test(value) || value === '') {
      setNamespaceValidation(true);
    } else {
      setNamespaceValidation(false);
    }
    setNamespace(value);
    profile.namespace = value;
    handleUpdateProfile(profile);
  };

  const handleChangeYaml = (event: ChangeEvent<HTMLTextAreaElement>) =>
    setYaml(event.target.value);

  const handleUpdateProfiles = useCallback(() => {
    profile.values.forEach(item => {
      if (item.version === version) {
        item.yaml = yaml;
      }
    });

    handleUpdateProfile(profile);

    setOpenYamlPreview(false);
  }, [profile, handleUpdateProfile, version, yaml]);

  useEffect(() => {
    const [selectedValue] = profile.values.filter(
      value => value.selected === true,
    );
    setNamespace(profile.namespace || '');
    if (selectedValue) {
      setVersion(selectedValue.version);
      setYaml(selectedValue.yaml);
    } else {
      setVersion(profile.values[0].version);
      setYaml(profile.values[0].yaml);
      profile.values[0].selected = true;
    }
  }, [profile]);

  return (
    <>
      <ListItemWrapper>
        <ListItem
          data-profile-name={profile.name}
          style={{ paddingLeft: '0px' }}
        >
          <div className="profile-version">
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
          <div className="profile-namespace">
            <FormControl>
              <Input
                id="profile-namespace"
                value={namespace}
                placeholder="flux-system"
                onChange={handleChangeNamespace}
                error={!isNamespaceValid}
              />
            </FormControl>
          </div>
          <Button
            style={{ marginLeft: medium }}
            variant="text"
            onClick={handleYamlPreview}
          >
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
          {loadingYaml ? (
            <Loader />
          ) : (
            <TextareaAutosize
              className={classes.textarea}
              defaultValue={yaml}
              onChange={event => handleChangeYaml(event)}
            />
          )}
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
