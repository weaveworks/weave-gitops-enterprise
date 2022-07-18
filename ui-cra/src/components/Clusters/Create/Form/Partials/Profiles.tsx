import React, { Dispatch, FC, useEffect, useState } from 'react';
import { UpdatedProfile } from '../../../../../types/custom';
import useProfiles from '../../../../../contexts/Profiles';
import styled from 'styled-components';
import { Loader } from '../../../../Loader';
import { DataTable } from '@weaveworks/weave-gitops';
import { Checkbox } from '@material-ui/core';
import { theme as weaveTheme } from '@weaveworks/weave-gitops';
import ProfilesListItem from './ProfileListItem';
import _ from 'lodash';

const ProfilesWrapper = styled.div`
  padding-bottom: ${({ theme }) => theme.spacing.xl};
  table {
    thead {
      th:first-of-type {
        padding: ${({ theme }) => theme.spacing.base};
      }
    }
    td:first-of-type {
      text-overflow: clip;
      width: 25px;
    }
    a {
      color: ${({ theme }) => theme.colors.primary};
    }
  }
`;

const Profiles: FC<{
  selectedProfiles: UpdatedProfile[];
  setSelectedProfiles: Dispatch<React.SetStateAction<UpdatedProfile[]>>;
}> = ({ selectedProfiles, setSelectedProfiles }) => {
  const { profiles, isLoading } = useProfiles();
  const [selected, setSelected] = useState<UpdatedProfile['name'][]>([]);

  const getProfilesFromNames = (names: string[]) =>
    profiles.filter(profile => names.find(name => profile.name === name));

  const getNamesFromProfiles = (profiles: UpdatedProfile[]) =>
    profiles.map(p => p.name);

  const handleIndividualClick = (
    event: React.ChangeEvent<HTMLInputElement>,
    name: string,
  ) => {
    if (event.target.checked === true) {
      const newProfilesNames = [...selected, name];
      setSelected(newProfilesNames);
      setSelectedProfiles(getProfilesFromNames(newProfilesNames));
    } else {
      const newProfilesNames = selected.filter(p => p !== name);
      setSelected(newProfilesNames);
      setSelectedProfiles(getProfilesFromNames(newProfilesNames));
    }
  };

  const handleSelectAllClick = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.checked) {
      setSelectedProfiles(profiles);
      return;
    }
    setSelectedProfiles([]);
  };

  const onlyRequiredItems =
    profiles.filter(item => item.required === true).length === profiles.length;
  const isAllSelected =
    profiles.length > 0 &&
    (selected.length === profiles.length || onlyRequiredItems);

  useEffect(() => {
    let requiredProfiles: UpdatedProfile[] = [];
    if (selectedProfiles.length === 0) {
      requiredProfiles = profiles.filter(profile => profile.required);
      setSelected(getNamesFromProfiles(requiredProfiles));
      setSelectedProfiles(requiredProfiles);
    }
    setSelected(getNamesFromProfiles(selectedProfiles));
  }, [selectedProfiles, profiles, setSelectedProfiles]);

  return isLoading ? (
    <Loader />
  ) : (
    <ProfilesWrapper>
      <h2>Profiles</h2>
      <DataTable
        rows={[
          ..._.differenceBy(profiles, selectedProfiles, 'name'),
          ...selectedProfiles,
        ]}
        fields={[
          {
            label: 'checkbox',
            labelRenderer: () => (
              <Checkbox
                onChange={handleSelectAllClick}
                checked={isAllSelected}
                indeterminate={
                  selected.length > 0 && selected.length < profiles.length
                }
                style={{
                  color: weaveTheme.colors.primary,
                }}
              />
            ),
            value: (profile: UpdatedProfile) => (
              <Checkbox
                onChange={event => handleIndividualClick(event, profile.name)}
                checked={selected.indexOf(profile.name) > -1}
                style={{
                  color: weaveTheme.colors.primary,
                }}
              />
            ),
            maxWidth: 25,
          },
          {
            label: 'Name',
            value: (p: UpdatedProfile) => (
              <span data-profile-name={p.name}>{p.name}</span>
            ),
            sortValue: ({ name }) => name,
            maxWidth: 275,
          },
          {
            label: 'Layer',
            value: (p: UpdatedProfile) =>
              p.layer ? (
                <div className="profile-layer">
                  <span>{p.layer}</span>
                </div>
              ) : null,
          },
          {
            label: 'Version',
            value: (p: UpdatedProfile) => (
              <ProfilesListItem
                profile={p}
                selectedProfiles={selectedProfiles}
                setSelectedProfiles={setSelectedProfiles}
              />
            ),
          },
        ]}
      />
    </ProfilesWrapper>
  );
};

export default Profiles;
