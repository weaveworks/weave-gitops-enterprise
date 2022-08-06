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
  width: 85%;
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
    .profile-details {
      display: flex;
      justify-content: space-around;
    }
  }
`;

const ProfileDetailsLabelRenderer = () => (
  <div className="profile-details">
    <h2>Version</h2>
    <h2>Namespace</h2>
    <h2>Yaml</h2>
  </div>
);

const Profiles: FC<{
  selectedProfiles: UpdatedProfile[];
  setSelectedProfiles: Dispatch<React.SetStateAction<UpdatedProfile[]>>;
}> = ({ selectedProfiles, setSelectedProfiles }) => {
  const getNamesFromProfiles = (profiles: UpdatedProfile[]) =>
    profiles.map(p => p.name);

  const { profiles, isLoading } = useProfiles();
  const [selected, setSelected] = useState<UpdatedProfile['name'][]>(
    getNamesFromProfiles(selectedProfiles),
  );

  const getProfilesFromNames = (names: string[]) =>
    profiles.filter(profile => names.find(name => profile.name === name));

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
      setSelected(getNamesFromProfiles(profiles));
      setSelectedProfiles(profiles);
      return;
    }
    setSelected([]);
    setSelectedProfiles([]);
  };

  const numSelected = selectedProfiles.length;
  const rowCount = profiles.length || 0;

  useEffect(() => {
    let requiredProfiles: UpdatedProfile[] = [];
    if (selectedProfiles.length === 0) {
      requiredProfiles = profiles.filter(profile => profile.required);
      setSelected(getNamesFromProfiles(requiredProfiles));
      setSelectedProfiles(requiredProfiles);
    }
  }, [profiles, setSelectedProfiles, selectedProfiles.length]);

  return isLoading ? (
    <Loader />
  ) : (
    <ProfilesWrapper>
      <h2>Profiles</h2>
      <DataTable
        className="profiles-table"
        rows={_.orderBy(
          [
            ..._.differenceBy(profiles, selectedProfiles, 'name'),
            ...selectedProfiles,
          ],
          ['name'],
          ['asc'],
        )}
        fields={[
          {
            label: 'checkbox',
            labelRenderer: () => (
              <Checkbox
                onChange={handleSelectAllClick}
                checked={rowCount > 0 && numSelected === rowCount}
                indeterminate={numSelected > 0 && numSelected < rowCount}
                style={{
                  color: weaveTheme.colors.primary,
                }}
              />
            ),
            value: (profile: UpdatedProfile) => (
              <Checkbox
                onChange={event => handleIndividualClick(event, profile.name)}
                checked={selected.indexOf(profile.name) > -1}
                disabled={profile.required}
                style={{
                  color: profile.required ? undefined : weaveTheme.colors.primary,
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
            maxWidth: 220,
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
            labelRenderer: () => <ProfileDetailsLabelRenderer />,
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
