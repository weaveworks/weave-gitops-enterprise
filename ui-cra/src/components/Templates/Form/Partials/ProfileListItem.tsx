import {
  FormControl,
  Input,
  MenuItem,
  TextField,
  createStyles,
  makeStyles,
} from '@material-ui/core';
import { Autocomplete } from '@material-ui/lab';
import { Button } from '@weaveworks/weave-gitops';
import { debounce } from 'lodash';
import React, {
  ChangeEvent,
  Dispatch,
  FC,
  useCallback,
  useEffect,
  useState,
} from 'react';
import styled from 'styled-components';
import {
  ClusterNamespacedName,
  RepositoryRef,
} from '../../../../cluster-services/cluster_services.pb';
import { ProfilesIndex, UpdatedProfile } from '../../../../types/custom';
import { DEFAULT_PROFILE_NAMESPACE } from '../../../../utils/config';
import ChartValuesDialog from './ChartValuesDialog';
import Profiles from './Profiles';
const semverValid = require('semver/functions/valid');
const semverValidRange = require('semver/ranges/valid');
const semverMaxSatisfying = require('semver/ranges/max-satisfying');
const semverCoerce = require('semver/functions/coerce');

const ProfileWrapper = styled.div`
  display: flex;
  justify-content: space-around;
`;

const ProfilesListItem: FC<{
  cluster?: ClusterNamespacedName;
  profile: UpdatedProfile;
  context?: string;
  setUpdatedProfiles: Dispatch<React.SetStateAction<ProfilesIndex>>;
  helmRepo: RepositoryRef;
}> = ({ profile, cluster, setUpdatedProfiles, helmRepo }) => {
  const [version, setVersion] = useState<string>('');
  const [yaml, setYaml] = useState<string>('');
  const [openYamlPreview, setOpenYamlPreview] = useState<boolean>(false);
  const [namespace, setNamespace] = useState<string>();
  const [isNamespaceValid, setNamespaceValidation] = useState<boolean>(true);
  const [inValidVersionErrorMessage, setInValidVersionErrorMessage] =
    useState<string>('');
  const [isValidVersion, setIsValidVersion] = useState<boolean>(true);

  const useStyles = makeStyles(() =>
    createStyles({
      autoComplete: {
        cursor: 'pointer',
        minWidth: '155px',
        overflow: 'hidden',
        minHeight: '1.1876em',
        marginRight: '24px',
        input: {
          padding: '10px',
        },
      },
    }),
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

  const validateVersion = (version: string) => {
    if (semverValid(version) || semverValidRange(version)) {
      setVersion(version);
    } else {
      setInValidVersionErrorMessage('The provided semver is invalid');
      setIsValidVersion(false);
    }
  };

  const handleSelectVersion = useCallback(
    (value: string, reason: string) => {
      setInValidVersionErrorMessage('');
      setIsValidVersion(true);
      validateVersion(value);

      profile.values.forEach(item =>
        item.selected === true ? (item.selected = false) : null,
      );
      const filteredVersions = profile.values.filter(
        item => item.version !== value,
      );
      const selectedVersion = profile.values.find(
        item => item.version === value,
      );

      if (selectedVersion) {
        profile.values = [
          ...filteredVersions,
          { ...selectedVersion, selected: true },
        ];
        setYaml(selectedVersion.yaml as string);
      } else {
        if (isValidVersion) {
          // check for validation  and setting the version
          profile.values.push({ version: value, selected: true, yaml: '' });
        }
      }
      handleUpdateProfile(profile);
    },
    [profile, handleUpdateProfile],
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
    setNamespace(value);
    profile.namespace = value;
    handleUpdateProfile(profile);
  };

  const handleChangeYaml = (event: ChangeEvent<HTMLTextAreaElement>) =>
    setYaml(event.target.value);

  const handleUpdateProfiles = useCallback(() => {
    profile.values.forEach(item => {
      if (item.version === version) {
        item.yaml = yaml;
      }
    });

    handleUpdateProfile(profile);

    setOpenYamlPreview(false);
  }, [profile, handleUpdateProfile, version, yaml]);

  useEffect(() => {
    const [selectedValue] = profile.values.filter(
      value => value.selected === true,
    );
    setNamespace(profile.namespace || '');
    if (selectedValue) {
      setVersion(selectedValue.version);
      setYaml(selectedValue.yaml);
    } else {
      if (profile.values.length > 0) {
        setVersion(profile.values?.[0]?.version);
        setYaml(profile.values?.[0]?.yaml);
        profile.values[0].selected = true;
      }
    }
  }, [profile]);

  // TODO: debounce function to reduce the number of created objects from profile version

  return (
    <>
      <ProfileWrapper data-profile-name={profile.name}>
        <div className="profile-version">
          <FormControl>
            <Autocomplete
              disabled={profile.required && profile.values.length === 1}
              disableClearable
              freeSolo
              className={classes.autoComplete}
              options={profile.values.map(option => option.version)}
              onChange={(event, newValue, reason) => {
                handleSelectVersion(newValue, reason);
              }}
              value={version}
              autoSelect
              renderInput={params => (
                <TextField
                  {...params}
                  variant="standard"
                  error={!isValidVersion}
                  helperText={
                    !!inValidVersionErrorMessage && inValidVersionErrorMessage
                  }
                />
              )}
            />
          </FormControl>
        </div>
        <div className="profile-namespace">
          <FormControl>
            <Input
              id="profile-namespace"
              value={namespace}
              placeholder={DEFAULT_PROFILE_NAMESPACE}
              onChange={handleChangeNamespace}
              error={!isNamespaceValid}
            />
          </FormControl>
        </div>
        <Button variant="text" onClick={handleYamlPreview}>
          Values.yaml
        </Button>
      </ProfileWrapper>

      {openYamlPreview && (
        <ChartValuesDialog
          yaml={yaml}
          cluster={cluster}
          profile={profile}
          version={version}
          onChange={handleChangeYaml}
          onSave={handleUpdateProfiles}
          onClose={() => setOpenYamlPreview(false)}
          helmRepo={helmRepo}
        />
      )}
    </>
  );
};

export default ProfilesListItem;
