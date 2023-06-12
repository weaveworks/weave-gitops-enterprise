import { Checkbox } from '@material-ui/core';
import { DataTable, Flex } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import React, { Dispatch, FC } from 'react';
import styled from 'styled-components';
import {
  ClusterNamespacedName,
  RepositoryRef,
} from '../../../../cluster-services/cluster_services.pb';
import { ProfilesIndex, UpdatedProfile } from '../../../../types/custom';
import { Loader } from '../../../Loader';
import ProfilesListItem from './ProfileListItem';

const ProfilesWrapper = styled.div`
  width: 85%;
  padding-bottom: ${({ theme }) => theme.spacing.xl};
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
  }
`;

const ProfileDetailsLabelRenderer = () => (
  <Flex>
    <h2>Version - Namespace - Yaml</h2>
  </Flex>
);

const Profiles: FC<{
  cluster?: ClusterNamespacedName;
  context?: string;
  updatedProfiles: ProfilesIndex;
  setUpdatedProfiles: Dispatch<React.SetStateAction<ProfilesIndex>>;
  isLoading: boolean;
  isProfilesEnabled?: string;
  helmRepo: RepositoryRef;
}> = ({
  context,
  cluster,
  updatedProfiles,
  setUpdatedProfiles,
  isLoading,
  isProfilesEnabled = 'true',
  helmRepo,
}) => {
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
      <>
        <h2>{context === 'app' ? 'Helm Releases' : 'Profiles'}</h2>
        {isLoading && <Loader />}
        {!isLoading && (
          <DataTable
            className="profiles-table"
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
                    onChange={event =>
                      handleIndividualClick(event, profile.name)
                    }
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
                label: 'Version',
                labelRenderer: () => <ProfileDetailsLabelRenderer />,
                value: (p: UpdatedProfile) => (
                  <ProfilesListItem
                    cluster={cluster}
                    context={context}
                    profile={p}
                    setUpdatedProfiles={setUpdatedProfiles}
                    helmRepo={helmRepo}
                  />
                ),
              },
            ]}
            hideSearchAndFilters={true}
          />
        )}
      </>
    </ProfilesWrapper>
  );
};

export default Profiles;
