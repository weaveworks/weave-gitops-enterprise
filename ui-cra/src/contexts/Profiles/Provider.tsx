import React, { FC, useCallback, useEffect, useState } from 'react';
import { Profile, Template, UpdatedProfile } from '../../types/custom';
import { request } from '../../utils/request';
import { Profiles } from './index';
import { useHistory } from 'react-router-dom';
import useNotifications from './../Notifications';
import useTemplates from './../Templates';

const ProfilesProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(true);
  const [profiles, setProfiles] = useState<Profile[]>([]);
  const [updatedProfiles, setUpdatedProfiles] = useState<UpdatedProfile[]>([]);
  const [optionalProfilesWithValues, setOptionalProfilesWithValues] = useState<
    UpdatedProfile[]
  >([]);
  const { setNotifications } = useNotifications();
  const { activeTemplate } = useTemplates();

  const history = useHistory();

  const profilesUrl = '/v1/profiles';

  const getProfile = (profileName: string) =>
    profiles.find(profile => profile.name === profileName) || null;

  const getDefaultProfiles = (template: Template) => {
    if (template?.annotations) {
      const annotations = Object.entries(template?.annotations);
      const defaultProfiles = [];
      for (const [key, value] of annotations) {
        if (key.includes('capi.weave.works/profile')) {
          defaultProfiles.push({ ...JSON.parse(value), required: true });
        }
      }
      return defaultProfiles;
    }
    return [];
  };

  const getProfiles = useCallback(() => {
    setLoading(true);
    request('GET', profilesUrl, {
      cache: 'no-store',
    })
      .then(res => setProfiles(res.profiles))
      .catch(err => {
        setNotifications([{ message: err.message, variant: 'danger' }]);
      })
      .finally(() => setLoading(false));
  }, [setNotifications]);

  const getProfileYaml = useCallback((profile: Profile) => {
    setLoading(true);
    const version =
      profile.availableVersions[profile.availableVersions.length - 1];
    return request('GET', `${profilesUrl}/${profile.name}/${version}/values`, {
      headers: {
        Accept: 'application/octet-stream',
      },
    }).finally(() => setLoading(false));
  }, []);

  const getProfileValues = useCallback(
    (optionalProfiles: Profile[]) => {
      const isProfileUpdated = (profile: Profile) =>
        optionalProfilesWithValues.filter(
          p =>
            p.name === profile.name &&
            p.version ===
              profile.availableVersions[profile.availableVersions.length - 1],
        ).length !== 0;

      // console.log(optionalProfiles);

      optionalProfiles.forEach(profile => {
        if (!isProfileUpdated(profile)) {
          getProfileYaml(profile)
            .then(data => {
              console.log(profile, data);
              console.log(
                'optionalProfilesWithValues',
                optionalProfilesWithValues,
              );
              setOptionalProfilesWithValues([
                ...optionalProfilesWithValues,
                {
                  name: profile.name,
                  version:
                    profile.availableVersions[
                      profile.availableVersions.length - 1
                    ],
                  values: data.message,
                  required: false,
                },
              ]);
            })
            // You can also do a functional update 'setOptionalProfilesWithValues(o => ...)
            .catch(error =>
              setNotifications([{ message: error.message, variant: 'danger' }]),
            );
        }
      });
      // console.log(optionalProfilesWithValues);
    },
    [getProfileYaml, setNotifications, optionalProfilesWithValues],
  );

  useEffect(() => {
    getProfiles();
    return history.listen(getProfiles);
  }, [history, getProfiles, activeTemplate]);

  useEffect(() => {
    const defaultProfiles =
      activeTemplate && getDefaultProfiles(activeTemplate);

    const validDefaultProfiles =
      defaultProfiles?.filter(profile =>
        profiles.find(
          p =>
            profile.name === p.name &&
            profile.version ===
              p.availableVersions[p.availableVersions.length - 1],
        ),
      ) || [];

    console.log('validDefaultProfiles', validDefaultProfiles);

    const optionalProfiles =
      profiles?.filter(
        profile =>
          !validDefaultProfiles.find(
            p =>
              profile.name === p.name &&
              profile.availableVersions[
                profile.availableVersions.length - 1
              ] === p.version,
          ),
      ) || [];

    getProfileValues(optionalProfiles);

    // console.log(optionalProfiles);

    // console.log(p);
    // console.log(validDefaultProfiles);

    setUpdatedProfiles([
      ...optionalProfilesWithValues,
      ...validDefaultProfiles,
    ]);
  }, [
    optionalProfilesWithValues,
    profiles,
    activeTemplate,
    getProfileYaml,
    setNotifications,
    getProfileValues,
    history,
  ]);

  console.log(updatedProfiles);

  return (
    <Profiles.Provider
      value={{
        profiles,
        loading,
        getProfile,
        getProfileYaml,
        updatedProfiles,
      }}
    >
      {children}
    </Profiles.Provider>
  );
};

export default ProfilesProvider;
