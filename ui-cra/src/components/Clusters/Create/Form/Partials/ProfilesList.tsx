import React, { FC, useCallback, useEffect, useState } from 'react';
import { UpdatedProfile } from '../../../../../types/custom';
import Box from '@material-ui/core/Box';
import List from '@material-ui/core/List';
import ProfileListItem from './ProfileListItem';

const ProfilesList: FC<{
  selectedProfiles: UpdatedProfile[];
  onProfilesUpdate: (profiles: UpdatedProfile[]) => void;
}> = ({ selectedProfiles, onProfilesUpdate }) => {
  const [enrichedProfiles, setEnrichedProfiles] =
    useState<UpdatedProfile[]>(selectedProfiles);

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

  useEffect(() => {
    setEnrichedProfiles(selectedProfiles);
  }, [selectedProfiles]);

  return (
    <Box display="flex">
      <List className="profiles-list">
        {enrichedProfiles?.map((profile, index) => (
          <ProfileListItem
            key={index}
            profile={profile}
            updateProfile={handleUpdateProfile}
          />
        ))}
      </List>
    </Box>
  );
};

export default ProfilesList;
