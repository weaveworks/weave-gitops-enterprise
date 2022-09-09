import React, { FC, useCallback, useMemo, useState } from 'react';
import {
  GitopsClusterEnriched,
  ListProfilesResponse,
  Profile,
  ProfilesIndex,
  TemplateEnriched,
  UpdatedProfile,
} from '../../types/custom';
import { request } from '../../utils/request';
import { Profiles } from './index';
import useNotifications from './../Notifications';
import { useQuery } from 'react-query';
import { Template } from '../../cluster-services/cluster_services.pb';
import _ from 'lodash';
import { maybeFromBase64 } from '../../utils/base64';
import {
  getCreateRequestAnnotation,
  maybeParseJSON,
} from '../../components/Clusters/Form/utils';

const profilesUrl = '/v1/profiles';

const getProfileLayer = (profiles: UpdatedProfile[], name: string) => {
  return profiles.find(p => p.name === name)?.layer;
};

const getDefaultProfiles = (template: Template, profiles: UpdatedProfile[]) => {
  const annotations = Object.entries(template?.annotations || {});
  const defaultProfiles: UpdatedProfile[] = [];
  for (const [key, value] of annotations) {
    if (key.includes('capi.weave.works/profile')) {
      const data = maybeParseJSON(value);
      if (data) {
        const { name, version, values, editable } = data;
        defaultProfiles.push({
          name,
          editable,
          values: [{ version, yaml: values || '', selected: false }],
          required: true,
          selected: true,
          layer: getProfileLayer(profiles, name),
        });
      }
    }
  }
  return defaultProfiles;
};

const getValidDefaultProfiles = (
  template: Template | null,
  initialProfiles: UpdatedProfile[],
) => {
  const defaultProfiles =
    template && getDefaultProfiles(template, initialProfiles);
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
};

const toUpdatedProfiles = (profiles?: Profile[]): UpdatedProfile[] => {
  const accumulator: UpdatedProfile[] = [];
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
};

const setVersionAndValuesFromTemplate = (
  profiles: UpdatedProfile[],
  template: TemplateEnriched,
) => {
  // get default / required profiles for the active template
  const validDefaultProfiles = getValidDefaultProfiles(template, profiles);

  // get the optional profiles by excluding the default profiles from the /v1/profiles response
  const optionalProfiles =
    profiles.filter(
      profile =>
        !validDefaultProfiles.find(
          p =>
            profile.name === p.name &&
            profile.values.map(value => value.version === p.values[0].version)
              .length !== 0,
        ),
    ) || [];

  return [...optionalProfiles, ...validDefaultProfiles];
};

interface Props {
  template?: TemplateEnriched;
  cluster?: GitopsClusterEnriched;
}

const setVersionAndValuesFromCluster = (
  profiles: UpdatedProfile[],
  cluster: GitopsClusterEnriched,
) => {
  const clusterData = getCreateRequestAnnotation(cluster);

  const profilesIndex = _.keyBy(profiles, 'name');

  let clusterProfiles: ProfilesIndex = {};
  if (clusterData?.values) {
    for (let clusterDataProfile of clusterData.values) {
      const profile = profilesIndex[clusterDataProfile.name!];
      if (profile) {
        clusterProfiles[clusterDataProfile.name!] = {
          ...profile,
          selected: true,
          namespace: clusterDataProfile.namespace!,
          // find our version, select it and overwrite the values
          values: profile.values.map(v =>
            v.version === clusterDataProfile.version
              ? {
                  ...v,
                  selected: true,
                  yaml: maybeFromBase64(clusterDataProfile.values!),
                }
              : v,
          ),
        };
      }
    }
  }

  return _.sortBy(
    Object.values({
      ...profilesIndex,
      ...clusterProfiles,
    }),
    'name',
  );
};

const mergeClusterAndTemplate = (
  data: ListProfilesResponse | undefined,
  template: TemplateEnriched | undefined,
  cluster: GitopsClusterEnriched | undefined,
) => {
  if (data?.code === 2) {
    return [];
  }
  let profiles = toUpdatedProfiles(data?.profiles);
  if (template) {
    profiles = setVersionAndValuesFromTemplate(profiles, template);
  }
  if (cluster) {
    profiles = setVersionAndValuesFromCluster(profiles, cluster);
  }
  return profiles;
};

const ProfilesProvider: FC<Props> = ({ template, cluster, children }) => {
  const [loading, setLoading] = useState<boolean>(true);
  const { setNotifications } = useNotifications();
  const [helmRepo, setHelmRepo] = useState<{
    name: string;
    namespace: string;
  }>({ name: '', namespace: '' });

  const getProfileYaml = useCallback((name: string, version: string) => {
    setLoading(true);
    return request('GET', `${profilesUrl}/${name}/${version}/values`, {
      headers: {
        Accept: 'application/octet-stream',
      },
    }).finally(() => setLoading(false));
  }, []);

  const onError = (error: Error) =>
    setNotifications([{ message: { text: error.message }, variant: 'danger' }]);

  const { isLoading, data } = useQuery<ListProfilesResponse, Error>(
    ['profiles', helmRepo?.name, helmRepo?.namespace],
    () =>
      request(
        'GET',
        helmRepo?.name !== '' && helmRepo?.name !== ''
          ? profilesUrl +
              `?helmRepoName=${helmRepo?.name}&helmRepoNamespace=${helmRepo?.namespace}`
          : profilesUrl,
      ),
    {
      onError,
    },
  );

  const profiles = useMemo(
    () => mergeClusterAndTemplate(data, template, cluster),
    [data, template, cluster],
  );

  return (
    <Profiles.Provider
      value={{
        loading,
        isLoading,
        helmRepo,
        setHelmRepo,
        profiles,
        getProfileYaml,
      }}
    >
      {children}
    </Profiles.Provider>
  );
};

export default ProfilesProvider;
