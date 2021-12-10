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
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
}> = ({ activeStep, setActiveStep, clickedStep, formData, setFormData }) => {
  const { updatedProfiles } = useProfiles();
  const [selectedProfiles, setSelectedProfiles] = useState<UpdatedProfile[]>(
    [],
  );

  const handleSelectProfiles = useCallback(
    (profiles: UpdatedProfile[]) => {
      setSelectedProfiles(profiles);
      setFormData((prevState: any) => ({ ...prevState, profiles }));
    },
    [setFormData],
  );

  useEffect(() => {
    const requiredProfiles = updatedProfiles.filter(
      profile => profile.required === true,
    );

    if (formData?.profiles?.length > 0) {
      handleSelectProfiles(formData?.profiles);
    } else {
      handleSelectProfiles(requiredProfiles);
    }
  }, [updatedProfiles, handleSelectProfiles, formData?.profiles]);

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
