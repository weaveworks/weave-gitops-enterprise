import React, { FC, useCallback, useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
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
import { Template } from '../../cluster-services/cluster_services.pb';

const ProfilesProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(true);
  const { setNotifications } = useNotifications();
  const { activeTemplate } = useTemplates();
  const [initialProfiles, setInitialProfiles] = useState<UpdatedProfile[]>([]);
  const [profiles, setProfiles] = useState<UpdatedProfile[]>([]);

  const profilesUrl = '/v1/profiles';

  const history = useHistory();

  const getProfileYaml = useCallback((name: string, version: string) => {
    setLoading(true);
    return request('GET', `${profilesUrl}/${name}/${version}/values`, {
      headers: {
        Accept: 'application/octet-stream',
      },
    }).finally(() => setLoading(false));
  }, []);

  const getProfileLayer = useCallback(
    (name: string) => {
      return initialProfiles.find(p => p.name === name)?.layer;
    },
    [initialProfiles],
  );

  const getDefaultProfiles = useCallback(
    (template: Template) => {
      if (template?.annotations) {
        const annotations = Object.entries(template?.annotations);
        const defaultProfiles: UpdatedProfile[] = [];
        for (const [key, value] of annotations) {
          if (key.includes('capi.weave.works/profile')) {
            const { name, version, values, editableValues } = JSON.parse(value);
            if (values !== undefined) {
              defaultProfiles.push({
                name,
                editableValues,
                values: [{ version, yaml: values, selected: false }],
                required: true,
                layer: getProfileLayer(name),
              });
            } else {
              defaultProfiles.push({
                name,
                editableValues,
                values: [
                  {
                    version,
                    yaml: '',
                    selected: false,
                  },
                ],
                required: true,
                layer: getProfileLayer(name),
              });
            }
          }
        }
        return defaultProfiles;
      }
      return [];
    },
    [getProfileLayer],
  );

  const getValidDefaultProfiles = useCallback(() => {
    const defaultProfiles =
      activeTemplate && getDefaultProfiles(activeTemplate);
    return (
      defaultProfiles?.filter(defaultProfile =>
        initialProfiles.find(
          profile =>
            defaultProfile.name === profile.name &&
            profile.values.map(
              value => value.version === defaultProfile.values[0].version,
            )?.length !== 0,
        ),
      ) || []
    );
  }, [activeTemplate, initialProfiles, getDefaultProfiles]);

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
    setInitialProfiles(getProfiles(data?.profiles));
  };

  const { isLoading } = useQuery<ListProfilesResponse, Error>(
    'profiles',
    () => request('GET', profilesUrl),
    {
      onSuccess,
      onError,
    },
  );

  useEffect(() => {
    // get default / required profiles for the active template
    const validDefaultProfiles = getValidDefaultProfiles();

    // get the optional profiles by excluding the default profiles from the /v1/profiles response
    const optionalProfiles =
      initialProfiles.filter(
        profile =>
          !validDefaultProfiles.find(
            p =>
              profile.name === p.name &&
              profile.values.map(value => value.version === p.values[0].version)
                .length !== 0,
          ),
      ) || [];

    setProfiles([...optionalProfiles, ...validDefaultProfiles]);
  }, [activeTemplate, history, getValidDefaultProfiles, initialProfiles]);

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
