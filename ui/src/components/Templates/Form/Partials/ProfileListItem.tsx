import { Box, FormControl, Input, MenuItem, Select } from '@material-ui/core';
import { Button, Flex } from '@weaveworks/weave-gitops';
import React, {
  ChangeEvent,
  Dispatch,
  FC,
  useCallback,
  useEffect,
  useState,
} from 'react';
import {
  ClusterNamespacedName,
  RepositoryRef,
} from '../../../../cluster-services/cluster_services.pb';
import { ProfilesIndex, UpdatedProfile } from '../../../../types/custom';
import { DEFAULT_PROFILE_NAMESPACE } from '../../../../utils/config';
import ChartValuesDialog from './ChartValuesDialog';

const ProfilesListItem: FC<{
  className?: string;
  cluster?: ClusterNamespacedName;
  profile: UpdatedProfile;
  context?: string;
  setUpdatedProfiles: Dispatch<React.SetStateAction<ProfilesIndex>>;
  helmRepo: RepositoryRef;
}> = ({ profile, cluster, setUpdatedProfiles, helmRepo, className }) => {
  const [version, setVersion] = useState<string>('');
  const [yaml, setYaml] = useState<string>('');
  const [openYamlPreview, setOpenYamlPreview] = useState<boolean>(false);
  const [namespace, setNamespace] = useState<string>('');
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
    (event: ChangeEvent<{ name?: string | undefined; value: unknown }>) => {
      const value = event.target.value as string;
      setVersion(value);

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
  );

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
      <Flex data-profile-name={profile.name} className={className}>
        <Box className="profile-version">
          <FormControl>
            <Select
              disabled={profile.required && profile.values.length === 1}
              value={version}
              onChange={handleSelectVersion}
              autoWidth
              label="Versions"
            >
              {profileVersions(profile)}
            </Select>
          </FormControl>
        </Box>
        <Box className="profile-namespace">
          <FormControl>
            <Input
              id="profile-namespace"
              value={namespace}
              placeholder={DEFAULT_PROFILE_NAMESPACE}
              onChange={handleChangeNamespace}
              error={!isNamespaceValid}
            />
          </FormControl>
        </Box>
        <Button variant="text" onClick={handleYamlPreview}>
          Values.yaml
        </Button>
      </Flex>

      {openYamlPreview && (
        <ChartValuesDialog
          yaml={yaml}
          cluster={cluster}
          profile={profile}
          version={version}
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
