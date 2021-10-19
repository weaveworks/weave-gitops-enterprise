import React, { FC, useCallback, useEffect, useState } from 'react';
import { Profile } from '../../types/custom';
import { request } from '../../utils/request';
import { Profiles } from './index';
import { useHistory } from 'react-router-dom';
import useNotifications from './../Notifications';

const ProfilesProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(true);
  const [profiles, setProfiles] = useState<Profile[]>([]);
  const [error, setError] = React.useState<string | null>(null);
  const [activeProfile, setActiveProfile] = useState<Profile | null>(null);
  const [profilePreview, setProfilePreview] = useState<string | null>(null);
  const { setNotifications } = useNotifications();

  const history = useHistory();

  const profilesUrl = '/v1/profiles';

  const getProfile = (profileName: string) =>
    profiles.find(profile => profile.name === profileName) || null;

  const getProfiles = useCallback(() => {
    setLoading(true);
    request('GET', profilesUrl, {
      cache: 'no-store',
    })
      .then(res => {
        setProfiles(res.profiles);
        setError(null);
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  const renderProfile = useCallback(
    data => {
      setLoading(true);
      request('POST', `${profilesUrl}/${activeProfile?.name}/render`, {
        body: JSON.stringify(data),
      })
        // {"renderedProfile":"apiVersion: cluster.x-k8s.io/v1alpha3\nkind: Cluster\nmetadata:\n  name: cls-name-oct18\n  namespace: default\nspec:\n"}
        .then(data => setProfilePreview(data.renderedProfile))
        .catch(err =>
          setNotifications([{ message: err.message, variant: 'danger' }]),
        )
        .finally(() => setLoading(false));
    },
    [activeProfile, setNotifications],
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
        error,
        setError,
        getProfile,
        profilePreview,
        setProfilePreview,
        renderProfile,
        activeProfile,
        setActiveProfile,
      }}
    >
      {children}
    </Profiles.Provider>
  );
};

export default ProfilesProvider;
