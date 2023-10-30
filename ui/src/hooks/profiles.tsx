import _ from 'lodash';
import { useContext } from 'react';
import { useQueries, useQuery } from 'react-query';
import { useDeepCompareMemo } from 'use-deep-compare';
import {
  GetConfigResponse,
  ListChartsForRepositoryResponse,
  RepositoryRef,
} from '../cluster-services/cluster_services.pb';
import { maybeParseJSON } from '../components/Templates/Form/utils';
import { EnterpriseClientContext } from '../contexts/EnterpriseClient';
import useNotifications from '../contexts/Notifications';
import {
  EnhancedRepositoryChart,
  GitopsClusterEnriched,
  ProfilesIndex,
  TemplateEnriched,
  UpdatedProfile,
} from '../types/custom';
import { maybeFromBase64 } from '../utils/base64';
import { formatError } from '../utils/formatters';

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
    helm_repository: {
      name: string;
      namespace: string;
    };
  }[];
}

const getProfileLayer = (profiles: UpdatedProfile[], name: string) => {
  return profiles.find(p => p.name === name)?.layer;
};

const getDefaultProfiles = (
  template: TemplateEnriched,
  profiles: UpdatedProfile[],
) => {
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
          layer: profile.layer || getProfileLayer(profiles, profile.name!),
          repoName: profile.sourceRef.name,
          repoNamespace: profile.sourceRef.namespace,
        } as UpdatedProfile),
    ) || [];

  return defaultProfiles;
};

const toUpdatedProfiles = (
  profiles?: EnhancedRepositoryChart[],
): UpdatedProfile[] => {
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
          repoName: profile.repoName,
          repoNamespace: profile.repoNamespace,
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
            profile.repoName === p.repoName &&
            profile.repoNamespace === p.repoNamespace &&
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
  const profilesIndex = _.keyBy(profiles, profile => {
    return `${profile.name}/${profile.repoName}/${profile.repoNamespace}`;
  });

  const clusterProfiles: ProfilesIndex = {};
  if (clusterData?.values) {
    for (const clusterDataProfile of clusterData.values) {
      const index = `${clusterDataProfile.name}/${clusterDataProfile.helm_repository.name}/${clusterDataProfile.helm_repository.namespace}`;
      const profile = profilesIndex[index || ''];
      if (profile) {
        clusterProfiles[index || ''] = {
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
    profile => {
      return `${profile.name}/${profile.repoName}/${profile.repoNamespace}`;
    },
  );
};

const mergeClusterAndTemplate = (
  data: ListChartsForRepositoryResponse[],
  template: TemplateEnriched | undefined,
  clusterData: AnnotationData,
) => {
  let profiles = data?.flatMap(d =>
    toUpdatedProfiles(d?.charts as EnhancedRepositoryChart[]),
  );
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
  helmReposRefs: RepositoryRef[],
) => {
  const { setNotifications } = useNotifications();

  const { api } = useContext(EnterpriseClientContext);

  const clusterData =
    cluster?.annotations?.['templates.weave.works/create-request'];

  const getConfigResponse = useQuery<GetConfigResponse, Error>('config', () =>
    api.GetConfig({}),
  );

  const hrQueries = helmReposRefs.map(helmRepo => {
    return {
      queryKey: [
        'profiles',
        helmRepo.name,
        helmRepo.namespace,
        helmRepo.cluster?.name,
        helmRepo.cluster?.namespace,
      ],
      queryFn: () =>
        api.ListChartsForRepository({
          repository: {
            name: helmRepo.name,
            namespace: helmRepo.namespace,
            cluster: helmRepo.cluster
              ? {
                  name: helmRepo.cluster?.name,
                  namespace: helmRepo.cluster?.namespace,
                }
              : { name: getConfigResponse?.data?.managementClusterName },
          },
        }),
      onError: (error: Error) => setNotifications(formatError(error)),
    };
  });

  const results = useQueries(hrQueries);

  const isLoading = results.some(query => query.isLoading);

  const data = results.map(
    result => result.data,
  ) as ListChartsForRepositoryResponse[];

  const profiles = useDeepCompareMemo(
    () =>
      mergeClusterAndTemplate(
        data,
        template,
        maybeParseJSON(clusterData || ''),
      ),
    [isLoading, data, template, clusterData],
  );

  return {
    isLoading,
    profiles,
  };
};

export default useProfiles;
