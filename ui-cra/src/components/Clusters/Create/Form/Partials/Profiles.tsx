import React, { Dispatch, FC, useCallback, useEffect, useState } from 'react';
import { UpdatedProfile } from '../../../../../types/custom';
import useProfiles from '../../../../../contexts/Profiles';
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
  profiles: any;
  setProfiles: Dispatch<React.SetStateAction<any>>;
}> = ({ activeStep, setActiveStep, clickedStep, profiles, setProfiles }) => {
  const { updatedProfiles } = useProfiles();
  const [selectedProfiles, setSelectedProfiles] = useState<UpdatedProfile[]>(
    [],
  );

  const handleSelectProfiles = useCallback(
    (selectProfiles: UpdatedProfile[]) => setProfiles(selectProfiles),
    [setProfiles],
  );

  useEffect(() => {
    setSelectedProfiles(profiles);
  }, [profiles]);

  return (
    <FormStep
      className="profiles"
      title="Profiles"
      active={activeStep === 'Profiles'}
      clicked={clickedStep === 'Profiles'}
      setActiveStep={setActiveStep}
    >
      {updatedProfiles.length > 0 ? (
        <ProfilesDropdown>
          <span>Select profiles:&nbsp;</span>

          <MultiSelectDropdown
            allItems={updatedProfiles}
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
