import {
  FormControl,
  Input,
  MenuItem,
  Select,
  TextField,
} from '@material-ui/core';
import { Autocomplete } from '@material-ui/lab';
import { Button } from '@weaveworks/weave-gitops';
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
const semverClean = require('semver/functions/clean')
const semverValid = require('semver/functions/valid')
const semverMaxSatisfying = require('semver/ranges/max-satisfying')

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

  const profileVersions = (profile: UpdatedProfile) => [
    ...profile.values.map((value, index) => {
      const { version } = value;
      return (
        <MenuItem key={index} value={version}>
          {version}
        </MenuItem>
      );
    }),
  ];

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
    (value: string | null) => {
      // const value = event.target.value as string;
      console.log(semverValid(value))
      console.log(semverMaxSatisfying(profile.values.map(item=>item.version),value))

      value && setVersion(value);
      console.log({ value });

      profile.values.forEach(item =>
        item.selected === true ? (item.selected = false) : null,
      );

      profile.values.forEach(item => {
        if (item.version === value) {
          item.selected = true;
          setYaml(item.yaml as string);
          return;
        }
      });

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
      // console.log({ yaml });
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
      <ProfileWrapper data-profile-name={profile.name}>
        <div className="profile-version">
          <FormControl>
            {/* <Select
              disabled={profile.required && profile.values.length === 1}
              value={version}
              onChange={handleSelectVersion}
              autoWidth
              label="Versions"
            >
              {profileVersions(profile)}
            </Select> */}
            <Autocomplete
              disabled={profile.required && profile.values.length === 1}
              id="free-solo-demo"
              freeSolo
              options={profile.values.map(option => option.version)}
              onChange={(event, newValue) => {
                handleSelectVersion(newValue);
              }}
           
              autoSelect
              renderInput={params => (
                <TextField
                  {...params}
                  label="freeSolo"
                  margin="normal"
                  variant="outlined"
                  error={!semverValid(semverClean(version))}
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
