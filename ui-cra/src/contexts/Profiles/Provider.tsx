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
import { maybeParseJSON } from '../../components/Clusters/Form/utils';

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
        const { name, version, values, editable, namespace } = data;
        defaultProfiles.push({
          name,
          editable,
          values: [
            { version: version || '', yaml: values || '', selected: false },
          ],
          required: true,
          selected: true,
          layer: getProfileLayer(profiles, name),
          namespace,
        });
      }
    }
  }
  return defaultProfiles;
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
  let defaultProfiles = getDefaultProfiles(template, profiles);

  // get the optional profiles by excluding the default profiles from the /v1/profiles response
  const optionalProfiles =
    profiles.filter(
      profile =>
        !defaultProfiles.find(
          p =>
            profile.name === p.name &&
            profile.values.map(value => value.version === p.values[0].version)
              .length !== 0,
        ),
    ) || [];

  // populate default profiles versions with those from initial profiles where they are missing
  defaultProfiles = defaultProfiles.map(defaultProfile => {
    if (defaultProfile.values[0].version === '') {
      defaultProfile.values =
        profiles?.find(
          optionalProfile => optionalProfile.name === defaultProfile.name,
        )?.values || [];
    }
    return defaultProfile;
  });

  return [...optionalProfiles, ...defaultProfiles];
};

interface Props {
  template?: TemplateEnriched;
  cluster?: GitopsClusterEnriched;
}

interface AnnotationData {
  commit_message: string;
  credentials: Credential;
  description: string;
  head_branch: string;
  parameter_values: { [key: string]: string };
  template_name: string;
  title: string;
  values: {
    name: string;
    selected: boolean;
    namespace: string;
    values: {
      version: string;
      yaml: string;
      selected?: boolean | undefined;
    }[];
    version: string;
  }[];
}

const setVersionAndValuesFromCluster = (
  profiles: UpdatedProfile[],
  clusterData: AnnotationData,
) => {
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
                  yaml: maybeFromBase64(
                    clusterDataProfile.values! as unknown as string,
                  ) as string,
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
  clusterData: AnnotationData,
) => {
  if (data?.code === 2) {
    return [];
  }
  let profiles = toUpdatedProfiles(data?.profiles);
  if (template) {
    profiles = setVersionAndValuesFromTemplate(profiles, template);
  }
  if (clusterData) {
    profiles = setVersionAndValuesFromCluster(profiles, clusterData);
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
  const clusterData =
    cluster?.annotations?.['templates.weave.works/create-request'];

  const getProfileYaml = useCallback(
    (name: string, version: string) => {
      const profilesYamlUrl = `${profilesUrl}/${name}/${version}/values`;
      setLoading(true);
      return request(
        'GET',
        helmRepo?.name !== '' && helmRepo?.name !== ''
          ? profilesYamlUrl +
              `?helmRepoName=${helmRepo?.name}&helmRepoNamespace=${helmRepo?.namespace}`
          : profilesYamlUrl,
        {
          headers: {
            Accept: 'application/octet-stream',
          },
        },
      ).finally(() => setLoading(false));
    },
    [helmRepo.name, helmRepo.namespace],
  );

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
    () =>
      mergeClusterAndTemplate(
        data,
        template,
        maybeParseJSON(clusterData || ''),
      ),
    [data, template, clusterData],
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
