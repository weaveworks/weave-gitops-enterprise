import React, { Dispatch, FC, useCallback, useEffect, useState } from 'react';
import { UpdatedProfile } from '../../../../../types/custom';
import MultiSelectDropdown from '../../../../MultiSelectDropdown';
import useProfiles from '../../../../../contexts/Profiles';
import ProfilesList from './ProfilesList';
import styled from 'styled-components';
import { Loader } from '../../../../Loader';
import { DataTable, theme } from '@weaveworks/weave-gitops';
import { Checkbox, makeStyles } from '@material-ui/core';
import { theme as weaveTheme } from '@weaveworks/weave-gitops';

const ProfilesWrapper = styled.div`
  padding-bottom: ${({ theme }) => theme.spacing.xl};
  .profiles-select {
    display: flex;
    align-items: center;
  }
`;

const Profiles: FC<{
  selectedProfiles: UpdatedProfile[];
  setSelectedProfiles: Dispatch<React.SetStateAction<UpdatedProfile[]>>;
}> = ({ selectedProfiles, setSelectedProfiles }) => {
  const { profiles, isLoading } = useProfiles();
  const [selected, setSelected] = useState<any[]>([]);

  const getProfileFromName = useCallback(
    name => profiles.filter(profile => profile.name === name),
    [profiles],
  );

  const getNamesFromProfiles = (profiles: UpdatedProfile[]) =>
    profiles.map(p => p.name);

  const handleIndividualClick = (
    event: React.ChangeEvent<HTMLInputElement>,
    name: string,
  ) => {
    console.log(event.target.checked);
    console.log(name);
    if (event.target.checked === true) {
      setSelected(prevState => [...prevState, name]);
    } else {
      setSelected(prevState => prevState.filter(p => p.name !== name));
    }
  };

  const useStyles = makeStyles(theme => ({
    formControl: {
      margin: theme.spacing(1),
      width: 300,
    },
    indeterminateColor: {
      color: weaveTheme.colors.primary,
    },
    downloadBtn: {
      color: weaveTheme.colors.primary,
      padding: '0px',
    },
  }));

  const classes = useStyles();

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

  useEffect(
    () => setSelected(getNamesFromProfiles(selectedProfiles)),
    [selectedProfiles],
  );

  return isLoading ? (
    <Loader />
  ) : (
    <ProfilesWrapper>
      <h2>Profiles</h2>
      {/* <div className="profiles-select"> */}
      {/* <span>Select profiles:&nbsp;</span> */}
      {/* <MultiSelectDropdown
          allItems={profiles}
          preSelectedItems={selectedProfiles}
          onSelectItems={handleSelectProfiles}
        />
      </div>
      <ProfilesList
        selectedProfiles={selectedProfiles}
        onProfilesUpdate={handleSelectProfiles}
      /> */}
      <DataTable
        key={profiles.length}
        rows={profiles}
        fields={[
          {
            label: 'checkbox',
            labelRenderer: () => (
              <Checkbox
                classes={{ indeterminate: classes.indeterminateColor }}
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
                checked={
                  profile.required === true ||
                  selected.indexOf(profile.name) > -1
                }
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
            label: 'Version',
            value: (p: UpdatedProfile) => (
              // MultiSelectDropdown needs to change from Profiles to Versions
              <MultiSelectDropdown
                allItems={profiles}
                preSelectedItems={[]}
                onSelectItems={setSelectedProfiles}
              />
            ),
          },
          {
            label: 'Layer',
            value: (p: UpdatedProfile) => (
              <span data-profile-name={p.layer}>{p.layer}</span>
            ),
          },
          {
            label: 'Namespace',
            value: 'namespace',
          },
        ]}
      />
    </ProfilesWrapper>
  );
};

export default Profiles;
