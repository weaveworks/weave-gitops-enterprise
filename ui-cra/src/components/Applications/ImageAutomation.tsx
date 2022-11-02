import { coreClient, DataTable, Flex, Kind } from '@weaveworks/weave-gitops';
import _ from 'lodash';

import { useQuery } from 'react-query';
import styled from 'styled-components';
import { Pipelines } from '../../api/pipelines/pipelines.pb';
import { ImageRepository } from '../../api/pipelines/types.pb';

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function useGetReconciledObjects(
  automationName: string,
  namespace: string,
  clusterName: string,
) {
  return useQuery([clusterName, namespace, automationName], () =>
    coreClient
      .GetReconciledObjects({
        automationName,
        namespace,
        kinds: [{ group: 'apps', version: 'v1', kind: 'Deployment' }],
        automationKind: Kind.Kustomization,
        clusterName,
      })
      .then(res => _.map(res.objects, o => JSON.parse(o.payload as string))),
  );
}

function useListImageAutomations() {
  return useQuery('imgs', () => Pipelines.ListImageAutomationObjects({}), {
    cacheTime: Infinity,
    staleTime: Infinity,
    retry: false,
  });
}

function ImageAutomation({ className, name, namespace, clusterName }: Props) {
  const { data } = useGetReconciledObjects(name, namespace, clusterName);
  const { data: automations } = useListImageAutomations();

  //   @ts-ignore
  const images = _.flatten(_.map(data, d => d.spec.template.spec.containers));

  return (
    <div className={className}>
      <Flex wide>
        <DataTable
          fields={[
            {
              label: 'K8s Object',
              value: (i: ImageRepository) => {
                const obj = _.find(data, d =>
                  _.includes(d.spec.template.spec.containers, i),
                );

                return obj.metadata.name;
              },
            },
            {
              label: 'Kind',
              value: (i: ImageRepository) => {
                const obj = _.find(data, d =>
                  _.includes(d.spec.template.spec.containers, i),
                );

                return obj.kind;
              },
            },
            { value: 'image', label: 'Tag' },
            {
              label: 'Repo Object',
              value: (i: ImageRepository) => {
                const raw = _.first(i?.image?.split(':'));

                const repo = _.find(automations?.imageRepos, { image: raw });

                return repo?.image || '';
              },
            },
          ]}
          rows={images}
        />
      </Flex>
    </div>
  );
}

export default styled(ImageAutomation).attrs({
  className: 'ImageAutomation',
})`
  width: 100%;
`;
