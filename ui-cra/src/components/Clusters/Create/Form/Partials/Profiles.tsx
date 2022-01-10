import React, { Dispatch, FC, useEffect, useState } from 'react';
import { UpdatedProfile } from '../../../../../types/custom';
import MultiSelectDropdown from '../../../../MultiSelectDropdown';
import { FormStep } from '../Step';
import ProfilesList from './ProfilesList';
import styled from 'styled-components';

const ProfilesDropdown = styled.div`
  display: flex;
  align-items: center;
`;

const Profiles: FC<{
  activeStep: string | undefined;
  setActiveStep: Dispatch<React.SetStateAction<string | undefined>>;
  clickedStep: string;
  profiles: UpdatedProfile[];
  setProfiles: Dispatch<React.SetStateAction<any>>;
}> = ({ activeStep, setActiveStep, clickedStep, profiles, setProfiles }) => {
  const [selectedProfiles, setSelectedProfiles] = useState<UpdatedProfile[]>(
    [],
  );

  const handleSelectProfiles = (selectProfiles: UpdatedProfile[]) =>
    setSelectedProfiles(selectProfiles);

  useEffect(() => {
    setSelectedProfiles(profiles.filter(profile => profile.required));
  }, [profiles]);

  return (
    <FormStep
      className="profiles"
      title="Profiles"
      active={activeStep === 'Profiles'}
      clicked={clickedStep === 'Profiles'}
      setActiveStep={setActiveStep}
    >
      {profiles.length > 0 ? (
        <ProfilesDropdown className="profiles-select">
          <span>Select profiles:&nbsp;</span>
          <MultiSelectDropdown
            allItems={profiles}
            preSelectedItems={selectedProfiles}
            onSelectItems={handleSelectProfiles}
          />
        </ProfilesDropdown>
      ) : (
        'No profiles available.'
      )}
      <ProfilesList
        selectedProfiles={selectedProfiles}
        onProfilesUpdate={handleSelectProfiles}
      />
    </FormStep>
  );
};

export default Profiles;
