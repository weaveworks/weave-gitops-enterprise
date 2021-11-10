import React, {
  ChangeEvent,
  FC,
  useCallback,
  useEffect,
  useState,
} from 'react';
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

const ProfilesList: FC<{
  selectedProfiles: UpdatedProfile[];
  onProfilesUpdate: (profiles: UpdatedProfile[]) => void;
}> = ({ selectedProfiles, onProfilesUpdate }) => {
  const classes = useStyles();
  const { loading } = useProfiles();
  const [currentProfile, setCurrentProfile] = useState<UpdatedProfile>();
  const [currentProfilePreview, setCurrentProfilePreview] = useState<string>();
  const [openYamlPreview, setOpenYamlPreview] = useState<boolean>(false);
  const [enrichedProfiles, setEnrichedProfiles] =
    useState<UpdatedProfile[]>(selectedProfiles);

  const handlePreview = (profile: UpdatedProfile) => {
    setCurrentProfile(profile);
    setCurrentProfilePreview(profile.values);
    setOpenYamlPreview(true);
  };

  const handleChange = useCallback(
    (event: ChangeEvent<HTMLTextAreaElement>) => {
      const currentProfileIndex = enrichedProfiles.findIndex(
        profile => profile.name === currentProfile?.name,
      );
      enrichedProfiles[currentProfileIndex].values = event.target.value;
    },
    [currentProfile, enrichedProfiles],
  );

  const handleUpdateProfiles = useCallback(() => {
    onProfilesUpdate(enrichedProfiles);
    setOpenYamlPreview(false);
  }, [onProfilesUpdate, enrichedProfiles]);

  useEffect(() => {
    setEnrichedProfiles(selectedProfiles);
  }, [selectedProfiles]);

  return (
    <>
      <Box sx={{ width: '100%', maxWidth: 360, bgcolor: 'background.paper' }}>
        <List>
          {enrichedProfiles.map((profile, index) => (
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
