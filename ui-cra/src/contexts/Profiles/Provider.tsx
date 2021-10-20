import React, { FC, useCallback, useEffect, useMemo, useState } from 'react';
import { Profile } from '../../types/custom';
import { request } from '../../utils/request';
import { Profiles } from './index';
import { useHistory } from 'react-router-dom';
import useNotifications from './../Notifications';

const ProfilesProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(true);
  const [profiles, setProfiles] = useState<Profile[]>([]);
  const [profilePreview, setProfilePreview] = useState<string | null>(null);
  const { setNotifications } = useNotifications();

  const history = useHistory();

  const profilesUrl = '/v1/profiles';

  const FAKE_PROFILES = useMemo(
    () => [{ name: 'Profile 1' }, { name: 'Profile 2' }, { name: 'Profile 3' }],
    [],
  );

  const FAKE_PROFILE_YAML =
    'apiVersion: cluster.x-k8s.io/v1alpha3\nkind: Cluster\nmetadata:\n  name: cls-name-oct18\n  namespace: default\nspec:\n';

  const getProfile = (profileName: string) =>
    profiles.find(profile => profile.name === profileName) || null;

  const getProfiles = useCallback(() => {
    setLoading(true);
    request('GET', profilesUrl, {
      cache: 'no-store',
    })
      .then(res => setProfiles(res.profiles))
      .catch(err => {
        setNotifications([{ message: err.message, variant: 'danger' }]);
        setProfiles(FAKE_PROFILES);
      })
      .finally(() => setLoading(false));
  }, [FAKE_PROFILES, setNotifications]);

  const renderProfile = useCallback(
    profileName => {
      setLoading(true);
      request('GET', `${profilesUrl}/${profileName}/render`)
        .then(data => setProfilePreview(data.renderedProfile))
        .catch(err => {
          setNotifications([{ message: err.message, variant: 'danger' }]);
          setProfilePreview(FAKE_PROFILE_YAML);
        })
        .finally(() => setLoading(false));
    },
    [setNotifications],
  );

  useEffect(() => {
    getProfiles();
    return history.listen(getProfiles);
  }, [history, getProfiles]);

  return (
    <Profiles.Provider
      value={{
        profiles,
        loading,
        getProfile,
        profilePreview,
        setProfilePreview,
        renderProfile,
      }}
    >
      {children}
    </Profiles.Provider>
  );
};

export default ProfilesProvider;
