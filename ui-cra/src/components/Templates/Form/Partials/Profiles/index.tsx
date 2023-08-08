import { Checkbox, MenuItem, TextField } from '@material-ui/core';
import { Autocomplete } from '@material-ui/lab';
import { DataTable, Flex } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import React, { Dispatch, FC, useState } from 'react';
import styled from 'styled-components';
import {
  ClusterNamespacedName,
  RepositoryRef,
} from '../../../../../cluster-services/cluster_services.pb';
import { ProfilesIndex, UpdatedProfile } from '../../../../../types/custom';
import { Loader } from '../../../../Loader';
import ProfilesListItem from './ProfileListItem';
import { CheckBoxOutlineBlank, CheckBox } from '@material-ui/icons';

const icon = <CheckBoxOutlineBlank fontSize="small" />;
const checkedIcon = <CheckBox fontSize="small" />;

const ProfilesWrapper = styled.div`
  padding-bottom: ${({ theme }) => theme.spacing.xl};
  .table-wrapper {
    max-height: 500px;
    overflow: scroll;
  }
  table {
    thead {
      th:first-of-type {
        padding: ${({ theme }) => theme.spacing.xs}
          ${({ theme }) => theme.spacing.base};
      }
      h2 {
        line-height: 1;
      }
    }
    td:first-of-type {
      text-overflow: clip;
      width: 25px;
      padding-left: ${({ theme }) => theme.spacing.base};
    }
    a {
      color: ${({ theme }) => theme.colors.primary};
    }
    .profile-details {
      justify-content: space-evenly;
    }
    .MuiFormControl-root {
      min-width: 150px;
      margin-right: 24px;
    }
  }
`;

const ProfileDetailsLabelRenderer = () => (
  <Flex className="profile-details">
    <h2>Version</h2>
    <h2>Namespace</h2>
    <h2>Yaml</h2>
  </Flex>
);

const Profiles: FC<{
  cluster?: ClusterNamespacedName;
  context?: string;
  updatedProfiles: ProfilesIndex;
  setUpdatedProfiles: Dispatch<React.SetStateAction<ProfilesIndex>>;
  isLoading: boolean;
  isProfilesEnabled?: string;
  helmRepos: RepositoryRef[];
}> = ({
  context,
  cluster,
  updatedProfiles,
  setUpdatedProfiles,
  isLoading,
  isProfilesEnabled = 'true',
  helmRepos,
}) => {
  const [selectedHelmRepositories, setSelectedHelmRepositories] = useState<
    RepositoryRef[]
  >([]);

  const handleIndividualClick = (
    event: React.ChangeEvent<HTMLInputElement>,
    name: string,
  ) => {
    setUpdatedProfiles(sp => ({
      ...sp,
      [name]: {
        ...sp[name],
        selected: event.target.checked,
      },
    }));
  };

  const handleSelectAllClick = (event: React.ChangeEvent<HTMLInputElement>) => {
    setUpdatedProfiles(sp =>
      _.mapValues(sp, p => ({
        ...p,
        selected: event.target.checked || p.required,
      })),
    );
  };

  const updatedProfilesList = _.sortBy(Object.values(updatedProfiles), [
    'name',
  ]);
  const numSelected = updatedProfilesList.filter(up => up.selected).length;
  const rowCount = updatedProfilesList.length || 0;

  return (
    <ProfilesWrapper>
      <h2>{context === 'app' ? 'Helm Releases' : 'Profiles'}</h2>
      <Autocomplete
        multiple
        id="helmrepositories-select"
        options={helmRepos}
        disableCloseOnSelect
        getOptionLabel={option => option.name as string}
        onChange={(event, selectedHelmRepos) =>
          setSelectedHelmRepositories(selectedHelmRepos)
        }
        renderOption={(option: RepositoryRef, { selected }) => (
          <li>
            <Checkbox
              icon={icon}
              checkedIcon={checkedIcon}
              style={{ marginRight: 8 }}
              checked={selected}
            />
            {option.name}
          </li>
        )}
        style={{ width: '100%' }}
        renderInput={params => (
          <TextField
            {...params}
            label="HelmRepositories"
            placeholder="Helm Repositories"
          />
        )}
      />

      {isLoading && <Loader />}
      {!isLoading && (
        <DataTable
          className="profiles-table table-wrapper"
          // show only profiles from selectedHelmRepositories
          rows={updatedProfilesList}
          fields={[
            {
              label: 'checkbox',
              labelRenderer: () => (
                <Checkbox
                  onChange={handleSelectAllClick}
                  checked={rowCount > 0 && numSelected === rowCount}
                  indeterminate={numSelected > 0 && numSelected < rowCount}
                  color="primary"
                />
              ),
              value: (profile: UpdatedProfile) => (
                <Checkbox
                  onChange={event => handleIndividualClick(event, profile.name)}
                  checked={Boolean(updatedProfiles[profile.name]?.selected)}
                  disabled={profile.required}
                  color={profile.required ? undefined : 'primary'}
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
            ...(context !== 'app'
              ? [
                  {
                    label: 'Layer',
                    value: (p: UpdatedProfile) =>
                      p.layer ? (
                        <div className="profile-layer">
                          <span>{p.layer}</span>
                        </div>
                      ) : null,
                  },
                ]
              : []),
            {
              label: 'Repository',
              value: (p: UpdatedProfile) => (
                <span data-profile-repository={p.repoName}>{p.repoName}</span>
              ),
              sortValue: ({ repoName }) => repoName,
              maxWidth: 220,
            },
            {
              label: 'Version',
              labelRenderer: () => <ProfileDetailsLabelRenderer />,
              value: (p: UpdatedProfile) => (
                <ProfilesListItem
                  className="profile-details"
                  cluster={cluster}
                  context={context}
                  profile={p}
                  setUpdatedProfiles={setUpdatedProfiles}
                  helmRepo={
                    helmRepos.find(
                      hr => hr.name === p.repoName,
                    ) as RepositoryRef
                  }
                />
              ),
            },
          ]}
          hideSearchAndFilters={true}
        />
      )}
    </ProfilesWrapper>
  );
};

export default Profiles;
