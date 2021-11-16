import React, { FC, useCallback, useEffect, useState } from 'react';
import { Profile, Template, UpdatedProfile } from '../../types/custom';
import { request } from '../../utils/request';
import { Profiles } from './index';
import { useHistory } from 'react-router-dom';
import useNotifications from './../Notifications';
import useTemplates from './../Templates';

const ProfilesProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(true);
  const [updatedProfiles, setUpdatedProfiles] = useState<UpdatedProfile[]>([]);
  const { setNotifications } = useNotifications();
  const { activeTemplate } = useTemplates();
  const [profilesWithValues, setProfilesWithValues] = useState<
    UpdatedProfile[]
  >([]);

  const history = useHistory();

  const profilesUrl = '/v1/profiles';

  const getDefaultProfiles = (template: Template) => {
    if (template?.annotations) {
      const annotations = Object.entries(template?.annotations);
      const defaultProfiles = [] as UpdatedProfile[];
      for (const [key, value] of annotations) {
        if (key.includes('capi.weave.works/profile')) {
          const { name, version, values } = JSON.parse(value);
          defaultProfiles.push({
            name,
            values: [{ version, yaml: values || '', selected: false }],
            required: true,
          });
        }
      }
      return defaultProfiles;
    }
    return [];
  };

  const getProfileYaml = useCallback((name: string, version: string) => {
    setLoading(true);
    return request('GET', `${profilesUrl}/${name}/${version}/values`, {
      headers: {
        Accept: 'application/octet-stream',
      },
    }).finally(() => setLoading(false));
  }, []);

  const getProfileValues = useCallback(
    (profiles: Profile[]) => {
      const profileRequests = profiles.flatMap(profile =>
        profile.availableVersions.map(async version => {
          const profileName = profile.name;
          const data = await getProfileYaml(profileName, version);
          return { name: profileName, version: version, payload: data };
        }),
      );
      Promise.all(profileRequests)
        .then(data => {
          const profiles = data.reduce(
            (
              accumulator: {
                name: string;
                values: { version: string; yaml: string }[];
                required: boolean;
              }[],
              profile,
            ) => {
              const profileName = accumulator.find(
                acc => acc.name === profile.name,
              );
              const value = {
                version: profile.version,
                yaml: profile.payload.message || '',
                selected: false,
              };

              if (profileName) {
                profileName.values.push(value);
              } else {
                accumulator.push({
                  name: profile.name,
                  values: [value],
                  required: false,
                });
              }
              return accumulator;
            },
            [],
          );
          setProfilesWithValues(profiles);
        })
        .catch(err =>
          setNotifications([{ message: err.message, variant: 'danger' }]),
        );
    },
    [getProfileYaml, setNotifications],
  );

  const getProfiles = useCallback(() => {
    setLoading(true);
    request('GET', profilesUrl, {
      cache: 'no-store',
    })
      .then(res => getProfileValues(res.profiles))
      .catch(err => {
        setNotifications([{ message: err.message, variant: 'danger' }]);
      })
      .finally(() => setLoading(false));
  }, [setNotifications, getProfileValues]);

  const getValidDefaultProfiles = useCallback(() => {
    const defaultProfiles =
      activeTemplate && getDefaultProfiles(activeTemplate);
    return (
      defaultProfiles?.filter(defaultProfile =>
        profilesWithValues.find(
          profile =>
            defaultProfile.name === profile.name &&
            profile.values.map(
              value =>
                (value.version as string) ===
                (defaultProfile.values[0].version as string),
            )?.length !== 0,
        ),
      ) || []
    );
  }, [activeTemplate, profilesWithValues]);

  useEffect(() => {
    getProfiles();
    return history.listen(getProfiles);
  }, [history, getProfiles, activeTemplate, getProfileValues]);

  useEffect(() => {
    // get default / required profiles for the active template
    const validDefaultProfiles = getValidDefaultProfiles();

    // get the optional profiles by excluding the default profiles from the /v1/profiles response
    const optionalProfiles =
      profilesWithValues?.filter(
        profile =>
          !validDefaultProfiles.find(
            p =>
              profile.name === p.name &&
              profile.values.map(value => value.version === p.values[0].version)
                .length !== 0,
          ),
      ) || [];

    setUpdatedProfiles([...optionalProfiles, ...validDefaultProfiles]);
  }, [
    profilesWithValues,
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
        updatedProfiles,
      }}
    >
      {children}
    </Profiles.Provider>
  );
};

export default ProfilesProvider;
