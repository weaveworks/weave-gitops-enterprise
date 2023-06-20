import {
  DataTable,
  Flex,
  FluxObject,
  RouterTab,
  SubRouterTabs,
  Text,
  YamlView,
} from '@weaveworks/weave-gitops';
import { GetExternalSecretResponse } from '../../../cluster-services/cluster_services.pb';
import { Routes } from '../../../utils/nav';
import { RowHeaders, SectionRowHeader } from '../../RowHeader';
import ListEvents from './Events/ListEvents';

const SecretDetailsTabs = ({
  clusterName,
  namespace,
  externalSecretName,
  secretDetails,
}: {
  clusterName: string;
  namespace: string;
  externalSecretName: string;
  secretDetails: GetExternalSecretResponse;
}) => {
  const path = Routes.SecretDetails;
  const secretObj = new FluxObject({
    payload: secretDetails?.yaml,
  });
  const props = secretObj.obj?.spec?.data?.map((d: any) => ({
    key: d.secretKey,
    value: d.remoteRef.property,
  }));
  const secretDetailsHeaders: Array<SectionRowHeader> = [
    {
      rowkey: 'External Secret',
      value: externalSecretName,
    },
    {
      rowkey: 'K8s Secret',
      value: secretDetails.secretName,
    },
    {
      rowkey: 'Cluster',
      value: clusterName,
    },
    {
      rowkey: 'Secret Store',
      value: secretDetails.secretStore,
    },
    {
      rowkey: 'Secret Store Type',
      value: secretDetails.secretStoreType,
    },
    {
      rowkey: 'Secret path',
      value: secretDetails.secretPath,
    },
    {
      rowkey: 'Version',
      value: secretDetails.version,
    },
    ...(!secretObj.obj?.spec?.data?.length
      ? [
          {
            rowkey: 'Properties',
            value: 'All Properties Included',
          },
        ]
      : []),
  ];
  return (
    <SubRouterTabs rootPath={`${path}/details`}>
      <RouterTab name="Details" path={`${path}/details`}>
        <Flex column wide gap="8">
          <RowHeaders rows={secretDetailsHeaders} />
          {secretObj.obj?.spec?.data.length && (
            <>
              <Text size="medium" semiBold>
                Properties
              </Text>
              <DataTable
                key={props?.length}
                rows={props}
                fields={[
                  {
                    label: 'PROPERTY',
                    value: 'value',
                  },
                  {
                    label: 'SECRET KEY',
                    value: 'key',
                  },
                ]}
              />
            </>
          )}
        </Flex>
      </RouterTab>
      <RouterTab name="Events" path={`${path}/events`}>
        <ListEvents
          involvedObject={{
            name: externalSecretName,
            namespace: namespace || '',
            kind: 'ExternalSecret',
          }}
          clusterName={clusterName}
        />
      </RouterTab>
      <RouterTab name="Yaml" path={`${path}/yaml`}>
        <YamlView
          yaml={secretObj.yaml}
          object={{
            kind: 'ExternalSecret',
            name: externalSecretName,
            namespace,
          }}
        />
      </RouterTab>
    </SubRouterTabs>
  );
};

export default SecretDetailsTabs;
