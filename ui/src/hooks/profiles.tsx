import {
  GetConfigResponse,
  ListChartsForRepositoryResponse,
  RepositoryChart,
  RepositoryRef,
  Template,
} from '../cluster-services/cluster_services.pb';
import { maybeParseJSON } from '../components/Templates/Form/utils';
import { EnterpriseClientContext } from '../contexts/EnterpriseClient';
import useNotifications from '../contexts/Notifications';
import {
  GitopsClusterEnriched,
  ProfilesIndex,
  TemplateEnriched,
  UpdatedProfile,
} from '../types/custom';
import { maybeFromBase64 } from '../utils/base64';
import { formatError } from '../utils/formatters';
import _ from 'lodash';
import { useContext, useMemo } from 'react';
import { useQuery } from 'react-query';

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
    values: string;
    version: string;
  }[];
}

const getProfileLayer = (profiles: UpdatedProfile[], name: string) => {
  return profiles.find(p => p.name === name)?.layer;
};

const getDefaultProfiles = (template: Template, profiles: UpdatedProfile[]) => {
  const defaultProfiles: UpdatedProfile[] =
    template.profiles?.map(
      profile =>
        ({
          ...profile,
          values: [
            {
              version: profile.version || '',
              yaml: profile.values || '',
              selected: false,
            },
          ],
          selected: true,
          layer: profile.layer || getProfileLayer(profiles, profile.name || ''),
        } as UpdatedProfile),
    ) || [];

  return defaultProfiles;
};

const toUpdatedProfiles = (profiles?: RepositoryChart[]): UpdatedProfile[] => {
  const accumulator: UpdatedProfile[] = [];
  profiles?.flatMap(profile =>
    profile.versions?.forEach(version => {
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
          name: profile.name || '',
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

const setVersionAndValuesFromCluster = (
  profiles: UpdatedProfile[],
  clusterData: AnnotationData,
) => {
  const profilesIndex = _.keyBy(profiles, 'name');

  const clusterProfiles: ProfilesIndex = {};
  if (clusterData?.values) {
    for (const clusterDataProfile of clusterData.values) {
      const profile = profilesIndex[clusterDataProfile.name || ''];
      if (profile) {
        clusterProfiles[clusterDataProfile.name || ''] = {
          ...profile,
          selected: true,
          namespace: clusterDataProfile.namespace!,
          // find our version, select it and overwrite the values
          values: profile.values.map(v =>
            v.version === clusterDataProfile.version
              ? {
                  ...v,
                  selected: true,
                  yaml: maybeFromBase64(clusterDataProfile.values || ''),
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
  data: ListChartsForRepositoryResponse | undefined,
  template: TemplateEnriched | undefined,
  clusterData: AnnotationData,
) => {
  let profiles = toUpdatedProfiles(data?.charts);
  if (template) {
    profiles = setVersionAndValuesFromTemplate(profiles, template);
  }
  if (clusterData) {
    profiles = setVersionAndValuesFromCluster(profiles, clusterData);
  }
  return profiles;
};

const useProfiles = (
  enabled: boolean,
  template: TemplateEnriched | undefined,
  cluster: GitopsClusterEnriched | undefined,
  helmRepo: RepositoryRef,
) => {
  const { setNotifications } = useNotifications();

  const { api } = useContext(EnterpriseClientContext);

  const clusterData =
    cluster?.annotations?.['templates.weave.works/create-request'];

  const onError = (error: Error) => setNotifications(formatError(error));

  const getConfigResponse = useQuery<GetConfigResponse, Error>('config', () =>
    api.GetConfig({}),
  );

  const { isLoading, data } = useQuery<ListChartsForRepositoryResponse, Error>(
    [
      'profiles',
      helmRepo.name,
      helmRepo.namespace,
      helmRepo.cluster?.name,
      helmRepo.cluster?.namespace,
    ],
    () =>
      api.ListChartsForRepository({
        repository: {
          name: helmRepo.name || 'weaveworks-charts',
          namespace: helmRepo.namespace || 'flux-system',
          cluster: helmRepo.cluster
            ? {
                name: helmRepo.cluster?.name,
                namespace: helmRepo.cluster?.namespace,
              }
            : { name: getConfigResponse?.data?.managementClusterName },
        },
      }),
    {
      enabled: enabled && !!getConfigResponse?.data?.managementClusterName,
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

  return {
    isLoading,
    profiles,
  };
};
export default useProfiles;
