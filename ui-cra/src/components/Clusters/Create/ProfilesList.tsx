import React, { FC, useCallback, useState } from 'react';
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

const ProfilesList: FC<{ selectedProfiles: Profile['name'][] }> = ({
  selectedProfiles,
}) => {
  const classes = useStyles();
  const { profilePreview, renderProfile, loading } = useProfiles();
  const [activeProfile, setActiveProfile] = useState<Profile['name']>();
  const [openYamlPreview, setOpenYamlPreview] = useState<boolean>(false);
  const rows = (profilePreview?.split('\n').length || 0) - 1;

  const handlePreview = useCallback(
    (event: any) => {
      console.log(event.target);
      setOpenYamlPreview(true);
      setActiveProfile(event.target.textContent);
      renderProfile(event.target.textContent);
    },
    [setOpenYamlPreview, renderProfile],
  );

  return (
    <>
      <Box sx={{ width: '100%', maxWidth: 360, bgcolor: 'background.paper' }}>
        <List>
          {selectedProfiles.map((profile, index) => {
            return (
              <ListItem key={index} onClick={handlePreview}>
                <ListItemText>{profile}</ListItemText>
                <ListItemText>values.yaml</ListItemText>
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
                />
                <OnClickAction
                  id="edit-yaml"
                  onClick={() => console.log('Call save yaml')}
                  text="Save changes"
                  className="success"
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
