import React, { ChangeEvent, FC, useCallback, useState } from 'react';
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

const ProfilesList: FC<{ selectedProfiles: Profile[] }> = ({
  selectedProfiles,
}) => {
  const classes = useStyles();
  const { profilePreview, renderProfile, loading } = useProfiles();
  const [activeProfile, setActiveProfile] = useState<Profile['name']>();
  const [openYamlPreview, setOpenYamlPreview] = useState<boolean>(false);
  const [updatedProfiles, setUpdatedProfiles] = useState(selectedProfiles);

  const rows = (profilePreview?.split('\n').length || 0) - 1;

  const handlePreview = useCallback(
    (profileName: string) => {
      setOpenYamlPreview(true);
      setActiveProfile(profileName);
      renderProfile(profileName);
    },
    [setOpenYamlPreview, renderProfile],
  );

  const handleChange = useCallback(
    (event: ChangeEvent<HTMLTextAreaElement>) => {
      console.log(event.target.value);
      console.log(activeProfile);
      // find the active profile in profileContent and update its state => setProfileContent(event.target.value);
    },
    [],
  );
  const handleUpdateProfiles = useCallback(() => {
    //save data from form textarea and send it up to the parent form to be submitted
  }, []);

  return (
    <>
      <Box sx={{ width: '100%', maxWidth: 360, bgcolor: 'background.paper' }}>
        <List>
          {selectedProfiles.map((profile, index) => {
            return (
              <ListItem key={index}>
                <ListItemText>{profile.name}</ListItemText>
                <Button
                  className={classes.downloadBtn}
                  onClick={() => handlePreview(profile.name)}
                >
                  values.yaml
                </Button>
              </ListItem>
            );
          })}
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
            <Typography variant="h5">{activeProfile}</Typography>
            <CloseIconButton onClick={() => setOpenYamlPreview(false)} />
          </DialogTitle>
          <DialogContent>
            {!loading ? (
              <>
                <textarea
                  className={classes.textarea}
                  rows={rows}
                  defaultValue={profilePreview || ''}
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
