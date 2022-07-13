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
  const onlyRequiredItems =
    profiles.filter(item => item.required === true).length === profiles.length;
  const isAllSelected =
    profiles.length > 0 &&
    (selected.length === profiles.length || onlyRequiredItems);

  const getItemsFromNames = (names: string[]) =>
    profiles.filter(item => names.find(name => item.name === name));

  const getNamesFromItems = (items: any[]) => items.map(item => item.name);

  const handleChange = (event: React.ChangeEvent<any>) => {
    const value = event.target.value;
    if (value[value.length - 1] === 'all') {
      const selectedItems = selected.length === profiles.length ? [] : profiles;
      setSelected(getNamesFromItems(selectedItems));
      // onSelectItems(selectedItems);
      return;
    }
    setSelected(value);
    // onSelectItems(getItemsFromNames(value));
  };

  const numSelected = selectedProfiles.length;
  const rowCount = profiles.length || 0;

  const handleSelectProfiles = (selectProfiles: UpdatedProfile[]) =>
    setSelectedProfiles(selectProfiles);

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

  useEffect(() => {
    if (selectedProfiles.length === 0) {
      setSelectedProfiles(profiles.filter(profile => profile.required));
    }
  }, [profiles, setSelectedProfiles, selectedProfiles.length]);

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
        key={selectedProfiles.length}
        rows={profiles}
        fields={[
          {
            label: 'checkbox',
            labelRenderer: () => (
              <Checkbox
                classes={{ indeterminate: classes.indeterminateColor }}
                checked={isAllSelected}
                indeterminate={
                  selected.length > 0 && selected.length < profiles.length
                }
                style={{
                  color: weaveTheme.colors.primary,
                }}
              />
            ),
            value: (p: UpdatedProfile) => (
              <Checkbox
                checked={p.required === true || selected.indexOf(p.name) > -1}
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
                preSelectedItems={selectedProfiles}
                onSelectItems={handleSelectProfiles}
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
