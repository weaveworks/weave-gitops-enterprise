import {
  FormControl,
  Input,
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
import { Tooltip } from '../../../Shared';
import ChartValuesDialog from './ChartValuesDialog';
const semverValid = require('semver/functions/valid');
const semverValidRange = require('semver/ranges/valid');
const semverMaxSatisfying = require('semver/ranges/max-satisfying');

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
  const [availableVersions] = useState(
    profile.values.map(item => item.version),
  );
  const useStyles = makeStyles(() =>
    createStyles({
      autoComplete: { minWidth: '155px', overflow: 'hidden' },
      input: {},
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

  const handleSelectVersion = useCallback(
    (value: string) => {
      setVersion(value);
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
      }
      handleUpdateProfile(profile);
    },
    [profile, handleUpdateProfile],
  );
  const handleInputChange = useCallback(
    (value: string) => {
      setInValidVersionErrorMessage('');
      setIsValidVersion(true);
      if ((semverValid(value) || semverValidRange(value)) && value !== '') {
        setVersion(value);

        const selectedVersion = profile.values.find(
          item => item.version === value,
        );

        if (!selectedVersion) {
          profile.values.forEach(item =>
            item.selected === true ? (item.selected = false) : null,
          );
          profile.values.push({ version: value, selected: true, yaml: '' });
        }
        handleUpdateProfile(profile);
      } else {
        setInValidVersionErrorMessage(
          'The provided version | range  is invalid',
        );
        setIsValidVersion(false);
      }
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
                options={profile.values.map(option => option.version)}
                onChange={(event, newValue) => {
                  handleSelectVersion(newValue);
                }}
                onInputChange={debounce(
                  (event, newInputValue) => handleInputChange(newInputValue),
                  500,
                )}
                value={version}
                renderInput={params => (
                  <TextField
                    {...params}
                    InputProps={{
                      ...params.InputProps,
                      className: classes.input,
                    }}
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
                disabled={!Boolean(profile.selected)}
              />
            </FormControl>
          </div>
          <Tooltip
            title="There is no Yaml file for this version | range"
            placement="top"
            disabled={Boolean(
              semverMaxSatisfying(availableVersions, version) && isValidVersion,
            )}
          >
            <div>
              <Button
                disabled={
                  !Boolean(
                    semverMaxSatisfying(availableVersions, version) &&
                      isValidVersion,
                  ) || !Boolean(profile.selected)
                }
                variant="text"
                onClick={handleYamlPreview}
              >
                Values.yaml
              </Button>
            </div>
          </Tooltip>
        </ProfileWrapper>
      </Tooltip>

      {openYamlPreview && (
        <ChartValuesDialog
          yaml={yaml}
          cluster={cluster}
          profile={profile}
          version={semverMaxSatisfying(availableVersions, version)}
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
