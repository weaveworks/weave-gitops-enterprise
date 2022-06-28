import React, { FC, useCallback, useState } from 'react';
import {
  ListProfilesResponse,
  Profile,
  UpdatedProfile,
} from '../../types/custom';
import { request } from '../../utils/request';
import { Profiles } from './index';
import useNotifications from './../Notifications';
import useTemplates from './../Templates';
import { useQuery } from 'react-query';

const ProfilesProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(true);
  const { setNotifications } = useNotifications();
  const { activeTemplate } = useTemplates();
  const [profiles, setProfiles] = useState<UpdatedProfile[]>([]);

  const profilesUrl = '/v1/profiles';

  const getProfileYaml = useCallback((name: string, version: string) => {
    setLoading(true);
    return request('GET', `${profilesUrl}/${name}/${version}/values`, {
      headers: {
        Accept: 'application/octet-stream',
      },
    }).finally(() => setLoading(false));
  }, []);

  const getProfilesWithDefaults = useCallback(
    (profiles: UpdatedProfile[]) => {
      if (activeTemplate?.annotations) {
        const annotations = Object.entries(activeTemplate?.annotations);
        for (const [key, value] of annotations) {
          if (key.includes('capi.weave.works/profile')) {
            const { name, version, values } = JSON.parse(value);
            profiles.forEach(profile => {
              const profileVersion = profile.values.find(
                value => value.version === version,
              )?.version;
              if (profile.name === name && profileVersion !== undefined) {
                profile.required = true;
                profile.values.forEach(value => {
                  if (value.version === version) {
                    value.yaml = values;
                  }
                });
              }
            });
          }
        }
      }
      return profiles;
    },
    [activeTemplate?.annotations],
  );

  const getProfiles = useCallback((profiles?: Profile[]) => {
    const accumulator: {
      name: string;
      values: { version: string; yaml: string; selected: boolean }[];
      required: boolean;
      layer?: string;
    }[] = [];
    profiles?.flatMap(profile =>
      profile.availableVersions.forEach(version => {
        const profileName = accumulator.find(p => p.name === profile.name);
        const value = {
          version,
          yaml: '',
          selected: false,
        };
        if (profileName) {
          profileName.values.push(value);
        } else {
          accumulator.push({
            name: profile.name,
            values: [value],
            required: false,
            layer: profile.layer,
          });
        }
      }),
    );
    return accumulator;
  }, []);

  const onError = (error: Error) =>
    setNotifications([{ message: { text: error.message }, variant: 'danger' }]);

  const onSuccess = (data: ListProfilesResponse) => {
    if (data.code === 2) {
      setProfiles([]);
      return;
    }
    setProfiles(getProfilesWithDefaults(getProfiles(data?.profiles)));
  };

  const { isLoading } = useQuery<ListProfilesResponse, Error>(
    'profiles',
    () => request('GET', profilesUrl),
    {
      onSuccess,
      onError,
    },
  );

  return (
    <Profiles.Provider
      value={{
        loading,
        isLoading,
        profiles,
        getProfileYaml,
      }}
    >
      {children}
    </Profiles.Provider>
  );
};

export default ProfilesProvider;
