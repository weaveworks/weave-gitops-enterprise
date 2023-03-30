import {
  FormControl,
  Input,
  TextField,
  createStyles,
  makeStyles,
} from '@material-ui/core';
import { Autocomplete, AutocompleteRenderInputParams } from '@material-ui/lab';
import { Button } from '@weaveworks/weave-gitops';
import { debounce } from 'lodash';
import React, {
  ChangeEvent,
  Dispatch,
  FC,
  useCallback,
  useMemo,
  useState,
} from 'react';
import styled from 'styled-components';
import {
  ClusterNamespacedName,
  RepositoryRef,
} from '../../../../cluster-services/cluster_services.pb';
import { ProfilesIndex, UpdatedProfile } from '../../../../types/custom';
import { DEFAULT_PROFILE_NAMESPACE } from '../../../../utils/config';
import { Tooltip } from '../../../Shared';
import ChartValuesDialog from './ChartValuesDialog';
const semverValid = require('semver/functions/valid');
const semverValidRange = require('semver/ranges/valid');
const semverMaxSatisfying = require('semver/ranges/max-satisfying');

const ProfileWrapper = styled.div`
  display: flex;
  justify-content: space-around;
`;

const useStyles = makeStyles(() =>
  createStyles({
    autoComplete: { minWidth: '155px', overflow: 'hidden' },
    input: {},
  }),
);

const ProfilesListItem: FC<{
  cluster?: ClusterNamespacedName;
  profile: UpdatedProfile;
  context?: string;
  setUpdatedProfiles: Dispatch<React.SetStateAction<ProfilesIndex>>;
  helmRepo: RepositoryRef;
}> = ({ profile, cluster, setUpdatedProfiles, helmRepo }) => {
  const [yaml, setYaml] = useState<string>('');
  const [openYamlPreview, setOpenYamlPreview] = useState<boolean>(false);
  const [isNamespaceValid, setNamespaceValidation] = useState<boolean>(true);
  const [inValidVersionErrorMessage, setInValidVersionErrorMessage] =
    useState<string>('');
  const [isValidVersion, setIsValidVersion] = useState<boolean>(true);
  const [availableVersions] = useState(
    profile.values.map(item => item.version),
  );

  const classes = useStyles();

  const handleUpdateProfile = useCallback(
    profile => {
      setUpdatedProfiles(sp => ({
        ...sp,
        [profile.name]: profile,
      }));
    },
    [setUpdatedProfiles],
  );

  const handleSelectVersion = useCallback(
    (event, value: string) => {
      const filteredVersions = profile.values.filter(
        item => item.version !== value,
      );
      const selectedVersion = profile.values.find(
        item => item.version === value,
      );

      if (selectedVersion) {
        profile.values.forEach(item =>
          item.selected === true ? (item.selected = false) : null,
        );
        profile.values = [
          ...filteredVersions,
          { ...selectedVersion, selected: true },
        ];
        setYaml(selectedVersion.yaml as string);
        handleUpdateProfile(profile);
      }
    },
    [profile, handleUpdateProfile],
  );

  const handleInputChange = useCallback(
    (event, value: string) => {
      setInValidVersionErrorMessage('');
      setIsValidVersion(true);
      if ((semverValid(value) || semverValidRange(value)) && value !== '') {
        const selectedVersion = profile.values.find(
          item => item.version === value,
        );

        if (!selectedVersion) {
          profile.values.forEach(item =>
            item.selected === true ? (item.selected = false) : null,
          );
          profile.values.push({ version: value, selected: true, yaml: '' });
          handleUpdateProfile(profile);
        }
      } else {
        setInValidVersionErrorMessage(
          'The provided version | range  is invalid',
        );
        setIsValidVersion(false);
      }
    },
    [profile, handleUpdateProfile],
  );

  const debouncedHandleInputChange = useMemo(
    () => debounce(handleInputChange, 500),
    [handleInputChange],
  );

  const handleYamlPreview = () => {
    setOpenYamlPreview(true);
  };

  const handleChangeNamespace = (event: ChangeEvent<HTMLInputElement>) => {
    const { value } = event.target;
    const pattern = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/;
    if (pattern.test(value) || value === '') {
      setNamespaceValidation(true);
    } else {
      setNamespaceValidation(false);
    }
    profile.namespace = value;
    handleUpdateProfile(profile);
  };

  const handleUpdateProfiles = useCallback(
    value => {
      setYaml(value);
      profile.values.forEach(item => {
        if (item.version === version) {
          item.yaml = value;
        }
      });

  const [selectedValue] = profile.values.filter(
    value => value.selected === true,
  );
  let version = '';
  if (selectedValue) {
    version = selectedValue.version;
  } else {
    version = profile.values?.[0].version || '';
  }

  const handleUpdateProfiles = useCallback(
    value => {
      setYaml(value);
      profile.values.forEach(item => {
        if (item.version === version) {
          item.yaml = value;
        }
      });

      handleUpdateProfile(profile);

      setOpenYamlPreview(false);
    },
    [profile, handleUpdateProfile, version],

  const handleRenderInput = useCallback(
    (params: AutocompleteRenderInputParams) => (
      <TextField
        {...params}
        InputProps={{
          ...params.InputProps,
          className: classes.input,
        }}
        variant="standard"
        error={!isValidVersion}
        helperText={!!inValidVersionErrorMessage && inValidVersionErrorMessage}
      />
    ),
    [classes.input, isValidVersion, inValidVersionErrorMessage],
  );

  const autocompleteVersions = useMemo(
    () => profile.values.map(option => option.version),
    [profile.values],
  );

  return (
    <>
      <Tooltip
        title="This fields are disabled as the profile is not checked"
        placement="top"
        disabled={Boolean(profile.selected)}
      >
        <ProfileWrapper data-profile-name={profile.name}>
          <div className="profile-version">
            <FormControl>
              <Autocomplete
                disabled={
                  (profile.required && profile.values.length === 1) ||
                  !Boolean(profile.selected)
                }
                disableClearable
                freeSolo
                className={classes.autoComplete}
                options={autocompleteVersions}
                onChange={handleSelectVersion}
                onInputChange={debouncedHandleInputChange}
                value={version}
                renderInput={handleRenderInput}
              />
            </FormControl>
          </div>
          <div className="profile-namespace">
            <FormControl>
              <Input
                id="profile-namespace"
                value={profile.namespace}
                placeholder={DEFAULT_PROFILE_NAMESPACE}
                onChange={handleChangeNamespace}
                error={!isNamespaceValid}
                disabled={!Boolean(profile.selected)}
              />
            </FormControl>
          </div>
          <div>
            <Button
              variant="text"
              onClick={handleYamlPreview}
              disabled={!Boolean(profile.selected)}
            >
              Values.yaml
            </Button>
          </div>
        </ProfileWrapper>
      </Tooltip>

      {openYamlPreview && (
        <ChartValuesDialog
          yaml={yaml}
          cluster={cluster}
          profile={profile}
          version={
            semverMaxSatisfying(availableVersions, version) ||
            availableVersions[0]
          }
          onSave={handleUpdateProfiles}
          onClose={() => setOpenYamlPreview(false)}
          helmRepo={helmRepo}
          onDiscard={() => setOpenYamlPreview(false)}
        />
      )}
    </>
  );
};

export default ProfilesListItem;
