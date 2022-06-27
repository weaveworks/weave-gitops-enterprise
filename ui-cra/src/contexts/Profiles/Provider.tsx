import React, { FC, useCallback, useEffect, useState } from 'react';
import {
  ListProfilesResponse,
  Profile,
  UpdatedProfile,
} from '../../types/custom';
import { request } from '../../utils/request';
import { Profiles } from './index';
import { useHistory } from 'react-router-dom';
import useNotifications from './../Notifications';
import useTemplates from './../Templates';
import { Template } from '../../cluster-services/cluster_services.pb';
import { useQuery } from 'react-query';

const ProfilesProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(true);
  const [updatedProfiles, setUpdatedProfiles] = useState<Profile[]>([]);
  const { setNotifications } = useNotifications();
  const { activeTemplate } = useTemplates();
  const [profilesWithValues, setProfilesWithValues] = useState<
    UpdatedProfile[]
  >([]);
  const [profiles, setProfiles] = useState<Profile[] | undefined>([]);

  const history = useHistory();

  const profilesUrl = '/v1/profiles';

  const getProfileYaml = useCallback((name: string, version: string) => {
    setLoading(true);
    return request('GET', `${profilesUrl}/${name}/${version}/values`, {
      headers: {
        Accept: 'application/octet-stream',
      },
    }).finally(() => setLoading(false));
  }, []);

  const getVersionValue = useCallback(
    (name: string, version: string) => {
      let versionValue: string = '';
      getProfileYaml(name, version).then(res => (versionValue = res));
      return versionValue;
    },
    [getProfileYaml],
  );

  const getProfileLayer = useCallback(
    (name: string) => {
      return profiles?.find(p => p.name === name)?.layer;
    },
    [profiles],
  );

  const getDefaultProfiles = useCallback(
    (template: Template) => {
      if (template?.annotations) {
        const annotations = Object.entries(template?.annotations);
        const defaultProfiles: UpdatedProfile[] = [];
        for (const [key, value] of annotations) {
          if (key.includes('capi.weave.works/profile')) {
            const { name, version, values } = JSON.parse(value);
            console.log(name, version, 'values', values);
            if (values !== undefined) {
              defaultProfiles.push({
                name,
                values: [{ version, yaml: values, selected: false }],
                required: true,
                layer: getProfileLayer(name),
              });
            } else {
              defaultProfiles.push({
                name,
                values: [
                  {
                    version,
                    yaml: getVersionValue(name, version),
                    selected: false,
                  },
                ],
                required: true,
                layer: getProfileLayer(name),
              });
            }
          }
        }
        console.log(defaultProfiles);

        return defaultProfiles;
      }
      return [];
    },
    [getVersionValue, getProfileLayer],
  );

  const getProfileValues = useCallback(
    (profiles: Profile[]) => {
      const profileRequests = profiles.flatMap(profile =>
        profile.availableVersions.map(async version => {
          const profileName = profile.name;
          const profileLayer = profile.layer;
          const data = await getProfileYaml(profileName, version);
          return {
            name: profileName,
            version: version,
            payload: data,
            layer: profileLayer,
          };
        }),
      );
      Promise.all(profileRequests)
        .then(data => {
          const profiles = data.reduce(
            (
              profilesWithAddedValues: {
                name: string;
                values: { version: string; yaml: string }[];
                required: boolean;
                layer?: string;
              }[],
              profile,
            ) => {
              const profileName = profilesWithAddedValues.find(
                p => p.name === profile.name,
              );
              const value = {
                version: profile.version,
                yaml: profile.payload.message || '',
                selected: false,
              };

              if (profileName) {
                profileName.values.push(value);
              } else {
                profilesWithAddedValues.push({
                  name: profile.name,
                  values: [value],
                  required: false,
                  layer: profile.layer,
                });
              }
              return profilesWithAddedValues;
            },
            [],
          );
          setProfilesWithValues(profiles);
        })
        .catch(err =>
          setNotifications([
            { message: { text: err.message }, variant: 'danger' },
          ]),
        );
    },
    [getProfileYaml, setNotifications],
  );

  const onError = (error: Error) =>
    setNotifications([{ message: { text: error.message }, variant: 'danger' }]);

  const onSuccess = (data: ListProfilesResponse) => {
    if (data.code === 2) {
      setUpdatedProfiles([]);
      return;
    }
    setProfiles(data?.profiles);
  };

  const { isLoading } = useQuery<ListProfilesResponse, Error>(
    'profiles',
    () => request('GET', profilesUrl),
    {
      onSuccess,
      onError,
    },
  );

  const getValidDefaultProfiles = useCallback(() => {
    const defaultProfiles =
      activeTemplate && getDefaultProfiles(activeTemplate);
    return (
      defaultProfiles?.filter(defaultProfile =>
        profiles?.find(
          profile =>
            defaultProfile.name === profile.name &&
            profile?.availableVersions.map(
              version => version === defaultProfile.values[0].version,
            )?.length !== 0,
        ),
      ) || []
    );
  }, [activeTemplate, profiles, getDefaultProfiles]);

  useEffect(() => {
    // get default / required profiles for the active template
    const validDefaultProfiles = getValidDefaultProfiles();

    // get the optional profiles by excluding the default profiles from the /v1/profiles response
    const optionalProfiles =
      profiles?.filter(
        profile =>
          !validDefaultProfiles.find(
            p =>
              profile.name === p.name &&
              profile.availableVersions.map(
                version => version === p.values[0].version,
              ).length !== 0,
          ),
      ) || [];
    console.log(optionalProfiles, validDefaultProfiles);
    setUpdatedProfiles([] as any);
  }, [
    profiles,
    activeTemplate,
    getProfileYaml,
    setNotifications,
    getProfileValues,
    history,
    getValidDefaultProfiles,
  ]);

  return (
    <Profiles.Provider
      value={{
        loading,
        isLoading,
        updatedProfiles,
      }}
    >
      {children}
    </Profiles.Provider>
  );
};

export default ProfilesProvider;
