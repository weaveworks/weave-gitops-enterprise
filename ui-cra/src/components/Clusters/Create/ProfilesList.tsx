import React, { ChangeEvent, Dispatch, FC, useCallback, useState } from 'react';
import ListItemText from '@material-ui/core/ListItemText';
import { makeStyles } from '@material-ui/core/styles';
import { Profile } from '../../../types/custom';
import Box from '@material-ui/core/Box';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import useProfiles from './../../../contexts/Profiles';
import Dialog from '@material-ui/core/Dialog';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';
import Typography from '@material-ui/core/Typography';
import { CloseIconButton } from '../../../assets/img/close-icon-button';
import { Loader } from '../../Loader';
import { OnClickAction } from '../../Action';
import weaveTheme from 'weaveworks-ui-components/lib/theme';
import Button from '@material-ui/core/Button';

const xs = weaveTheme.spacing.xs;

const useStyles = makeStyles(theme => ({
  formControl: {
    margin: theme.spacing(1),
    width: 300,
  },
  indeterminateColor: {
    color: '#00B3EC',
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
    color: '#00B3EC',
    padding: '0px',
  },
}));

const ProfilesList: FC<{
  selectedProfiles: Profile[];
  onProfilesUpdate: Dispatch<
    React.SetStateAction<
      { name: Profile['name']; version: string; values: string }[] | undefined
    >
  >;
}> = ({ selectedProfiles, onProfilesUpdate }) => {
  const classes = useStyles();
  const { renderProfile, loading } = useProfiles();
  const [currentProfile, setCurrentProfile] = useState<Profile>();
  const [currentProfilePreview, setCurrentProfilePreview] = useState<string>();
  const [openYamlPreview, setOpenYamlPreview] = useState<boolean>(false);
  const [updatedProfiles, setUpdatedProfiles] = useState<
    { name: Profile['name']; version: string; values: string }[]
  >([]);

  const rows = (currentProfilePreview?.split('\n').length || 0) - 1;

  const handlePreview = useCallback(
    (profile: Profile) => {
      setOpenYamlPreview(true);
      setCurrentProfile(profile);
      if (updatedProfiles.filter(p => p.name === profile.name).length === 0) {
        renderProfile(profile).then(data => {
          setCurrentProfilePreview(data.message);
          setUpdatedProfiles([
            ...updatedProfiles,
            {
              name: profile.name,
              version:
                profile.availableVersions[profile.availableVersions.length - 1],
              values: data.message,
            },
          ]);
        });
      } else {
        setCurrentProfilePreview(
          updatedProfiles?.find(p => p.name === profile.name)?.values,
        );
      }
    },
    [setOpenYamlPreview, renderProfile, updatedProfiles],
  );

  const handleChange = useCallback(
    (event: ChangeEvent<HTMLTextAreaElement>) => {
      const currentProfileIndex = updatedProfiles.findIndex(
        profile => profile.name === currentProfile?.name,
      );
      updatedProfiles[currentProfileIndex].values = event.target.value;
    },
    [currentProfile, updatedProfiles],
  );

  const handleUpdateProfiles = useCallback(() => {
    onProfilesUpdate(updatedProfiles);
    setOpenYamlPreview(false);
  }, [onProfilesUpdate, updatedProfiles]);

  return (
    <>
      <Box sx={{ width: '100%', maxWidth: 360, bgcolor: 'background.paper' }}>
        <List>
          {selectedProfiles.map((profile, index) => (
            <ListItem key={index}>
              <ListItemText>{profile.name}</ListItemText>
              <Button
                className={classes.downloadBtn}
                onClick={() => handlePreview(profile)}
              >
                values.yaml
              </Button>
            </ListItem>
          ))}
        </List>
      </Box>
      <Dialog
        open={openYamlPreview}
        maxWidth="md"
        fullWidth
        onClose={() => setOpenYamlPreview(false)}
      >
        <div id="preview-yaml-popup" className={classes.dialog}>
          <DialogTitle disableTypography>
            <Typography variant="h5">{currentProfile?.name}</Typography>
            <CloseIconButton onClick={() => setOpenYamlPreview(false)} />
          </DialogTitle>
          <DialogContent>
            {!loading ? (
              <>
                <textarea
                  className={classes.textarea}
                  rows={rows}
                  defaultValue={currentProfilePreview || ''}
                  onChange={handleChange}
                />
                <OnClickAction
                  id="edit-yaml"
                  onClick={handleUpdateProfiles}
                  text="Save changes"
                />
              </>
            ) : (
              <Loader />
            )}
          </DialogContent>
        </div>
      </Dialog>
    </>
  );
};

export default ProfilesList;
