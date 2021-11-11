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
import Box from '@material-ui/core/Box';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import useProfiles from '../../../../../contexts/Profiles';
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogActions,
  TextareaAutosize,
} from '@material-ui/core';
import Typography from '@material-ui/core/Typography';
import { CloseIconButton } from '../../../../../assets/img/close-icon-button';
import { Loader } from '../../../../Loader';
import { OnClickAction } from '../../../../Action';
import weaveTheme from 'weaveworks-ui-components/lib/theme';
import Button from '@material-ui/core/Button';
import { GitOpsBlue } from '../../../../../muiTheme';
import { Dropdown, DropdownItem } from 'weaveworks-ui-components';

const large = weaveTheme.spacing.large;
const medium = weaveTheme.spacing.medium;
const base = weaveTheme.spacing.base;
const xxs = weaveTheme.spacing.xxs;
const xs = weaveTheme.spacing.xs;

const useStyles = makeStyles(theme => ({
  formControl: {
    margin: theme.spacing(1),
    width: 300,
  },
  indeterminateColor: {
    color: GitOpsBlue,
  },
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

const ProfilesList: FC<{
  selectedProfiles: UpdatedProfile[];
  onProfilesUpdate: (profiles: UpdatedProfile[]) => void;
}> = ({ selectedProfiles, onProfilesUpdate }) => {
  const classes = useStyles();
  const { loading } = useProfiles();
  const [currentProfile, setCurrentProfile] = useState<UpdatedProfile>(
    {} as UpdatedProfile,
  );
  const [currentProfileVersion, setCurrentProfileVersion] = useState<string>();
  const [currentProfilePreview, setCurrentProfilePreview] = useState<string>();
  const [openYamlPreview, setOpenYamlPreview] = useState<boolean>(false);
  const [enrichedProfiles, setEnrichedProfiles] =
    useState<UpdatedProfile[]>(selectedProfiles);

  const profileVersions = (profile: UpdatedProfile) =>
    profile.values.map(value => {
      const { version } = value;
      return {
        label: version,
        value: version,
      };
    });

  const handlePreview = (profile: UpdatedProfile) => {
    setCurrentProfile(profile);
    // setCurrentProfilePreview(profile.values);
    setOpenYamlPreview(true);
  };

  const handleChange = useCallback(
    (event: ChangeEvent<HTMLTextAreaElement>) => {
      const currentProfileIndex = enrichedProfiles.findIndex(
        profile => profile.name === currentProfile?.name,
      );

      // name: "observability"
      // required: true
      // values: Array(1)
      // 0:
      // version: "0.0.1"
      // yaml: undefined
      // [[Prototype]]: Object
      // length: 1
      // [[Prototype]]: Array(0)

      // need to look for the version of the profile =>
      // enrichedProfiles[currentProfileIndex].values.find(value => currentProfileVersion === value.version).yaml
      // OLD: enrichedProfiles[currentProfileIndex].values = event.target.value;
    },
    [currentProfile, enrichedProfiles],
  );

  const handleSelectVersion = useCallback(
    (
      profile: UpdatedProfile,
      event: FormEvent<HTMLInputElement>,
      value: string,
    ) => {
      console.log(value);
      console.log(profile);
    },
    [],
  );

  const handleUpdateProfiles = useCallback(() => {
    onProfilesUpdate(enrichedProfiles);
    setOpenYamlPreview(false);
  }, [onProfilesUpdate, enrichedProfiles]);

  useEffect(() => {
    setEnrichedProfiles(selectedProfiles);
  }, [selectedProfiles]);

  console.log(enrichedProfiles);

  return (
    <>
      <Box sx={{ width: '100%', maxWidth: 360, bgcolor: 'background.paper' }}>
        <List>
          {enrichedProfiles.map((profile, index) => (
            <ListItemWrapper>
              <ListItem key={index} className="">
                <ListItemText className="profile-name">
                  {profile.name}
                </ListItemText>
                <div className="profile-version">
                  <span>Version</span>
                  <Dropdown
                    value={''}
                    disabled={loading}
                    items={profileVersions(profile)}
                    onChange={(event, value) =>
                      handleSelectVersion(profile, event, value)
                    }
                  />
                </div>
                <Button
                  className={classes.downloadBtn}
                  onClick={() => handlePreview(profile)}
                >
                  values.yaml
                </Button>
              </ListItem>
            </ListItemWrapper>
          ))}
        </List>
      </Box>
      <Dialog
        open={openYamlPreview}
        className={classes.dialog}
        maxWidth="md"
        fullWidth
        scroll="paper"
        onClose={() => setOpenYamlPreview(false)}
      >
        <DialogTitle disableTypography>
          <Typography variant="h5">{currentProfile?.name}</Typography>
          <CloseIconButton onClick={() => setOpenYamlPreview(false)} />
        </DialogTitle>
        <DialogContent>
          {!loading ? (
            <TextareaAutosize
              className={classes.textarea}
              defaultValue={currentProfilePreview || ''}
              onChange={handleChange}
            />
          ) : (
            <Loader />
          )}
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

export default ProfilesList;
