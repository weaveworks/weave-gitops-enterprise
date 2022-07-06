import React, { Dispatch, FC, useEffect } from 'react';
import { UpdatedProfile } from '../../../../../types/custom';
import MultiSelectDropdown from '../../../../MultiSelectDropdown';
import useProfiles from '../../../../../contexts/Profiles';
import ProfilesList from './ProfilesList';
import styled from 'styled-components';
import { Loader } from '../../../../Loader';

const ProfilesDropdown = styled.div`
  display: flex;
  align-items: center;
`;

const Profiles: FC<{
  selectedProfiles: UpdatedProfile[];
  setSelectedProfiles: Dispatch<React.SetStateAction<UpdatedProfile[]>>;
}> = ({ selectedProfiles, setSelectedProfiles }) => {
  const { profiles, isLoading } = useProfiles();

  const handleSelectProfiles = (selectProfiles: UpdatedProfile[]) =>
    setSelectedProfiles(selectProfiles);

  useEffect(() => {
    if (selectedProfiles.length === 0) {
      setSelectedProfiles(profiles.filter(profile => profile.required));
    }
  }, [profiles, setSelectedProfiles, selectedProfiles.length]);

  return isLoading ? (
    <Loader />
  ) : (
    <>
      <ProfilesDropdown className="profiles-select">
        <span>Select profiles:&nbsp;</span>
        <MultiSelectDropdown
          allItems={profiles}
          preSelectedItems={selectedProfiles}
          onSelectItems={handleSelectProfiles}
        />
      </ProfilesDropdown>
      <ProfilesList
        selectedProfiles={selectedProfiles}
        onProfilesUpdate={handleSelectProfiles}
      />
    </>
  );
};

export default Profiles;
