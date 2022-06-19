import React, { Dispatch, FC, useEffect } from 'react';
import { UpdatedProfile } from '../../../../../types/custom';
import MultiSelectDropdown from '../../../../MultiSelectDropdown';
import useProfiles from '../../../../../contexts/Profiles';
import { FormStep } from '../Step';
import ProfilesList from './ProfilesList';
import styled from 'styled-components';
import { Loader } from '../../../../Loader';

const ProfilesDropdown = styled.div`
  display: flex;
  align-items: center;
`;

const Profiles: FC<{
  activeStep: string | undefined;
  setActiveStep: Dispatch<React.SetStateAction<string | undefined>>;
  clickedStep: string;
  selectedProfiles: UpdatedProfile[];
  setSelectedProfiles: Dispatch<React.SetStateAction<UpdatedProfile[]>>;
}> = ({
  activeStep,
  setActiveStep,
  clickedStep,
  selectedProfiles,
  setSelectedProfiles,
}) => {
  const { updatedProfiles, loading } = useProfiles();

  const handleSelectProfiles = (selectProfiles: UpdatedProfile[]) =>
    setSelectedProfiles(selectProfiles);

  useEffect(() => {
    if (selectedProfiles.length === 0) {
      setSelectedProfiles(updatedProfiles.filter(profile => profile.required));
    }
  }, [updatedProfiles, setSelectedProfiles, selectedProfiles.length]);

  return (
    <FormStep
      className="profiles"
      title="Profiles"
      active={activeStep === 'Profiles'}
      clicked={clickedStep === 'Profiles'}
      setActiveStep={setActiveStep}
    >
      {loading ? (
        <Loader />
      ) : (
        <>
          <ProfilesDropdown className="profiles-select">
            <span>Select profiles:&nbsp;</span>
            <MultiSelectDropdown
              allItems={updatedProfiles}
              preSelectedItems={selectedProfiles}
              onSelectItems={handleSelectProfiles}
            />
          </ProfilesDropdown>
          <ProfilesList
            selectedProfiles={selectedProfiles}
            onProfilesUpdate={handleSelectProfiles}
          />
        </>
      )}
    </FormStep>
  );
};

export default Profiles;
