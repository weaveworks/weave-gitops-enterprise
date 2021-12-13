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
  const [selectedProfiles, setSelectedProfiles] =
    useState<UpdatedProfile[]>(profiles);

  console.log(profiles);

  const handleSelectProfiles = useCallback(
    (selectProfiles: UpdatedProfile[]) => {
      setSelectedProfiles(selectProfiles);
      // setSelectedProfiles((prevState: any) => ({
      //   ...prevState,
      //   profiles: selectProfiles,
      // }));
      setProfiles((prevState: any) => ({
        ...prevState,
        profiles: selectProfiles,
      }));
    },
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
      <ProfilesDropdown>
        <span>Select profiles:&nbsp;</span>
        <MultiSelectDropdown
          items={updatedProfiles}
          onSelectItems={handleSelectProfiles}
        />
      </ProfilesDropdown>
      <ProfilesList
        selectedProfiles={selectedProfiles}
        onProfilesUpdate={handleSelectProfiles}
      />
    </FormStep>
  );
};

export default Profiles;
