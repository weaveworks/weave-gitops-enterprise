import React, { FC, useMemo, useState } from 'react';
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
import { faTrashAlt } from '@fortawesome/free-solid-svg-icons';

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
}));

const FAKE_PROFILE_YAML =
  'apiVersion: cluster.x-k8s.io/v1alpha3\nkind: Cluster\nmetadata:\n  name: cls-name-oct18\n  namespace: default\nspec:\n';

const ProfilesList: FC<{ selectedProfiles: Profile[] }> = ({
  selectedProfiles,
}) => {
  const classes = useStyles();
  const { renderProfile, loading } = useProfiles();
  const [openYamlPreview, setOpenYamlPreview] = useState<boolean>(false);
  const rows = (FAKE_PROFILE_YAML?.split('\n').length || 0) - 1;

  return useMemo(() => {
    return (
      <Box sx={{ width: '100%', maxWidth: 360, bgcolor: 'background.paper' }}>
        <List>
          {selectedProfiles.map((profile, index) => {
            const yaml = '';
            // renderProfile(profile.name);
            return (
              <ListItem key={index}>
                <ListItemText>{profile.name}</ListItemText>
                <ListItemText
                  onClick={() => {
                    setOpenYamlPreview(true);
                    // get the textarea content for the specific profile
                  }}
                >
                  Values.yaml
                </ListItemText>
                {openYamlPreview && (
                  <Dialog
                    open
                    maxWidth="md"
                    fullWidth
                    onClose={() => console.log('')}
                  >
                    <div id="preview-yaml-popup" className={classes.dialog}>
                      <DialogTitle disableTypography>
                        <Typography variant="h5">
                          {profile.name} values.yaml
                        </Typography>
                        <CloseIconButton
                          onClick={() => setOpenYamlPreview(false)}
                        />
                      </DialogTitle>
                      <DialogContent>
                        {!loading ? (
                          <>
                            <textarea
                              className={classes.textarea}
                              rows={rows}
                              value={FAKE_PROFILE_YAML}
                              readOnly
                            />
                            <OnClickAction
                              id="edit-yaml"
                              onClick={() => console.log('Call save yaml')}
                              text="Edit profile"
                              className="success"
                            />
                          </>
                        ) : (
                          <Loader />
                        )}
                      </DialogContent>
                    </div>
                  </Dialog>
                )}
              </ListItem>
            );
          })}
        </List>
      </Box>
    );
  }, [
    renderProfile,
    selectedProfiles,
    classes,
    loading,
    openYamlPreview,
    rows,
  ]);
};

export default ProfilesList;
