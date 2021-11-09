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
  const { setNotifications } = useNotifications();
  const { activeTemplate } = useTemplates();
  const [profilesWithValues, setProfilesWithValues] = useState<
    UpdatedProfile[]
  >([]);

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
    (profiles: Profile[]) => {
      const isProfileUpdated = (profile: Profile) =>
        profilesWithValues.filter(
          p =>
            p.name === profile.name &&
            p.version ===
              profile.availableVersions[profile.availableVersions.length - 1],
        ).length !== 0;

      const profileRequests = [] as Promise<any>[];

      profiles.forEach(profile => {
        if (!isProfileUpdated(profile)) {
          profileRequests.push(
            getProfileYaml(profile)
              .then(data =>
                setProfilesWithValues([
                  ...profilesWithValues,
                  {
                    name: profile.name,
                    version:
                      profile.availableVersions[
                        profile.availableVersions.length - 1
                      ],
                    values: data.message,
                    required: false,
                  },
                ]),
              )
              .catch(error =>
                setNotifications([
                  { message: error.message, variant: 'danger' },
                ]),
              ),
          );
        }
      });
      return Promise.all(profileRequests);
    },
    [getProfileYaml, setNotifications, profilesWithValues],
  );

  const getValidDefaultProfiles = useCallback(() => {
    const defaultProfiles =
      activeTemplate && getDefaultProfiles(activeTemplate);
    return (
      defaultProfiles?.filter(profile =>
        profiles.find(
          p =>
            profile.name === p.name &&
            profile.version ===
              p.availableVersions[p.availableVersions.length - 1],
        ),
      ) || []
    );
  }, [activeTemplate, profiles]);

  useEffect(() => {
    getProfiles();
    return history.listen(getProfiles);
  }, [history, getProfiles, activeTemplate]);

  useEffect(() => {
    // get yaml values for all the profiles
    getProfileValues(profiles);

    // get default / required profiles for the active template
    const validDefaultProfiles = getValidDefaultProfiles();

    // get the optional profiles by excluding the default profiles from the /v1/profiles response
    const optionalProfiles =
      profilesWithValues?.filter(
        profile =>
          !validDefaultProfiles.find(
            p => profile.name === p.name && profile.version === p.version,
          ),
      ) || [];

    setUpdatedProfiles([...optionalProfiles, ...validDefaultProfiles]);
  }, [
    profilesWithValues,
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
