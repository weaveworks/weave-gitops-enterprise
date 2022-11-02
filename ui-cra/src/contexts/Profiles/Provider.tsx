import _ from 'lodash';
import { FC, useContext, useMemo, useState } from 'react';
import { useQuery } from 'react-query';
import {
  ListChartsForRepositoryResponse,
  RepositoryChart,
  Template,
} from '../../cluster-services/cluster_services.pb';
import { maybeParseJSON } from '../../components/Clusters/Form/utils';
import {
  GitopsClusterEnriched,
  ProfilesIndex,
  TemplateEnriched,
  UpdatedProfile,
} from '../../types/custom';
import { maybeFromBase64 } from '../../utils/base64';
import { EnterpriseClientContext } from '../EnterpriseClient';
import { Profiles } from './index';

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
          name: profile.name!,
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
    values: string;
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

const ProfilesProvider: FC<Props> = ({ template, cluster, children }) => {
  const [helmRepo, setHelmRepo] = useState<{
    name: string;
    namespace: string;
    clusterName: string;
    clusterNamespace: string;
  }>({ name: '', namespace: '', clusterName: '', clusterNamespace: '' });

  const { api } = useContext(EnterpriseClientContext);

  const clusterData =
    cluster?.annotations?.['templates.weave.works/create-request'];

  const { isLoading, data, error } = useQuery<
    ListChartsForRepositoryResponse,
    Error
  >(
    [
      'profiles',
      helmRepo.name,
      helmRepo.namespace,
      helmRepo.clusterName,
      helmRepo.clusterNamespace,
    ],
    () =>
      api.ListChartsForRepository({
        repository: {
          name: helmRepo.name || 'weaveworks-charts',
          namespace: helmRepo.namespace || 'flux-system',
          cluster: helmRepo.clusterName
            ? {
                name: helmRepo.clusterName,
                namespace: helmRepo.clusterNamespace,
              }
            : { name: 'management' },
        },
      }),
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
        isLoading,
        error,
        helmRepo,
        setHelmRepo,
        profiles,
      }}
    >
      {children}
    </Profiles.Provider>
  );
};

export default ProfilesProvider;
