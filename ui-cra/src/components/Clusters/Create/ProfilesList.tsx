import React, { FC, useState } from 'react';
import ListItemText from '@material-ui/core/ListItemText';
import { makeStyles } from '@material-ui/core/styles';
import { Profile } from '../../../types/custom';
import Box from '@material-ui/core/Box';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import useProfiles from './../../../contexts/Profiles';

const useStyles = makeStyles(theme => ({
  formControl: {
    margin: theme.spacing(1),
    width: 300,
  },
  indeterminateColor: {
    color: '#00B3EC',
  },
}));

const ProfilesList: FC<{ selectedProfiles: Profile[] }> = ({
  selectedProfiles,
}) => {
  const classes = useStyles();
  const { renderProfile } = useProfiles();

  return (
    <Box sx={{ width: '100%', maxWidth: 360, bgcolor: 'background.paper' }}>
      <List>
        {selectedProfiles.map(profile => (
          <ListItem>
            <ListItemText>
              {profile.name}
              {renderProfile(profile.name)}
            </ListItemText>
          </ListItem>
        ))}
      </List>
    </Box>
  );
};

export default ProfilesList;
