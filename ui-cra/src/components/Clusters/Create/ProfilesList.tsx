import React, {
  ChangeEvent,
  Dispatch,
  FC,
  useCallback,
  useEffect,
  useState,
} from 'react';
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
import useNotifications from './../../../contexts/Notifications';

const FAKE_PROFILE_YAML =
  'apiVersion: cluster.x-k8s.io/v1alpha3\nkind: Cluster\nmetadata:\n  name: cls-name-oct18\n  namespace: default\nspec:\n';

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
  const { profilePreview, renderProfile, loading } = useProfiles();
  const [activeProfile, setActiveProfile] = useState<Profile>();
  const [openYamlPreview, setOpenYamlPreview] = useState<boolean>(false);
  const [updatedProfiles, setUpdatedProfiles] = useState<
    { name: Profile['name']; version: string; values: string }[]
  >([]);
  const { setNotifications } = useNotifications();

  const rows = (profilePreview?.split('\n').length || 0) - 1;

  const handlePreview = useCallback(
    (profile: Profile) => {
      setOpenYamlPreview(true);
      setActiveProfile(profile);
      if (updatedProfiles.filter(p => p.name === profile.name).length === 0) {
        renderProfile(profile).then(data => {
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
      }
    },
    [setOpenYamlPreview, renderProfile, updatedProfiles],
  );

  const handleChange = useCallback(
    (event: ChangeEvent<HTMLTextAreaElement>) => {
      const currentProfileIndex = updatedProfiles.findIndex(
        profile => profile.name === activeProfile?.name,
      );
      updatedProfiles[currentProfileIndex].values = event.target.value;
    },
    [activeProfile, updatedProfiles],
  );

  const handleUpdateProfiles = useCallback(() => {
    onProfilesUpdate(updatedProfiles);
    setOpenYamlPreview(false);
  }, [onProfilesUpdate, updatedProfiles]);

  console.log(updatedProfiles);

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
            <Typography variant="h5">{activeProfile?.name}</Typography>
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
