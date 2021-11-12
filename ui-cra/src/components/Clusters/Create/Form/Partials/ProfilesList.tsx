import React, {
  ChangeEvent,
  FC,
  FormEvent,
  useCallback,
  useEffect,
  useState,
} from 'react';
import styled from 'styled-components';
import { makeStyles } from '@material-ui/core/styles';
import { Profile, UpdatedProfile } from '../../../../../types/custom';
import Box from '@material-ui/core/Box';
import List from '@material-ui/core/List';
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
import { GitOpsBlue } from '../../../../../muiTheme';
import ProfileListItem from './ProfileListItem';

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
}));

const ProfilesList: FC<{
  selectedProfiles: UpdatedProfile[];
  onProfilesUpdate: (profiles: UpdatedProfile[]) => void;
}> = ({ selectedProfiles, onProfilesUpdate }) => {
  const classes = useStyles();
  const [enrichedProfiles, setEnrichedProfiles] = useState<UpdatedProfile[]>(
    [],
  );

  const handleUpdateProfile = useCallback(
    profile => {
      const currentProfileIndex = enrichedProfiles.findIndex(
        p => p.name === profile.name,
      );

      enrichedProfiles[currentProfileIndex] = profile;

      onProfilesUpdate(enrichedProfiles);
    },
    [onProfilesUpdate, enrichedProfiles],
  );

  const selectedVersionYaml = (profile: UpdatedProfile) =>
    profile.values.find(value => value.selected === true);

  useEffect(() => {
    setEnrichedProfiles(selectedProfiles);
  }, [selectedProfiles]);

  return (
    <Box sx={{ width: '100%', maxWidth: 360, bgcolor: 'background.paper' }}>
      <List>
        {enrichedProfiles.map((profile, index) => {
          console.log(profile);
          const selected = selectedVersionYaml(profile);
          console.log(selected);
          return (
            <ProfileListItem
              key={index}
              profile={profile}
              updateProfile={handleUpdateProfile}
            />
          );
        })}
      </List>
    </Box>
  );
};

export default ProfilesList;
