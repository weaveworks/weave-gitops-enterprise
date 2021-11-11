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
          const { name, version, values } = JSON.parse(value);
          defaultProfiles.push({
            name,
            values: [{ version, yaml: values }],
            required: true,
          });
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

  const getProfileYaml = useCallback(
    (name: Profile['name'], version: string) => {
      setLoading(true);
      return request('GET', `${profilesUrl}/${name}/${version}/values`, {
        headers: {
          Accept: 'application/octet-stream',
        },
        referrer: `${name}/${version}`,
      }).finally(() => setLoading(false));
    },
    [],
  );

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
        .then(data => console.log(data))
        .catch(error => console.log(error));
    },
    [getProfileYaml],
  );

  const getValidDefaultProfiles = useCallback(() => {
    const defaultProfiles =
      activeTemplate && getDefaultProfiles(activeTemplate);
    return (
      defaultProfiles?.filter(defaultProfile =>
        profiles.find(
          profile =>
            defaultProfile.name === profile.name &&
            profile.availableVersions.find(
              availableVersion =>
                availableVersion === defaultProfile.values[0].version,
            )?.length !== 0,
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
            p =>
              profile.name === p.name &&
              profile.values.map(value => value.version === p.values[0].version)
                .length !== 0,
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
        updatedProfiles,
      }}
    >
      {children}
    </Profiles.Provider>
  );
};

export default ProfilesProvider;
