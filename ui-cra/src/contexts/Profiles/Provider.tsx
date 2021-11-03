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
  const [defaultProfiles, setDefaultProfiles] = useState<UpdatedProfile[]>([]);
  const [updatedProfiles, setUpdatedProfiles] = useState<UpdatedProfile[]>([]);
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
          defaultProfiles.push(JSON.parse(value));
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

  useEffect(() => {
    getProfiles();
    activeTemplate && setDefaultProfiles(getDefaultProfiles(activeTemplate));
    return history.listen(getProfiles);
  }, [history, getProfiles, activeTemplate]);

  useEffect(() => {
    const isProfileUpdated = (profile: Profile) =>
      updatedProfiles.filter(p => p.name === profile.name).length !== 0;
    profiles.forEach(profile => {
      if (!isProfileUpdated(profile)) {
        getProfileYaml(profile)
          .then(data => {
            setUpdatedProfiles([
              ...updatedProfiles,
              {
                name: profile.name,
                version:
                  profile.availableVersions[
                    profile.availableVersions.length - 1
                  ],
                values: data.message,
              },
            ]);
          })
          .catch(error =>
            setNotifications([{ message: error.message, variant: 'danger' }]),
          );
      }
    });
  }, [getProfileYaml, profiles, setNotifications, updatedProfiles]);

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
