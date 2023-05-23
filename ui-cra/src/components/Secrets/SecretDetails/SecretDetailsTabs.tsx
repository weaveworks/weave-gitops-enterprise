import { RouterTab, SubRouterTabs, YamlView } from '@weaveworks/weave-gitops';
import { GetExternalSecretResponse } from '../../../cluster-services/cluster_services.pb';
import styled from 'styled-components';
import { generateRowHeaders, SectionRowHeader } from '../../RowHeader';
import { Routes } from '../../../utils/nav';
import ListEvents from './Events/ListEvents';

const YAML = require('yaml');

const ListEventsWrapper = styled.div`
  width: 100%;
`;
const DetailsHeadersWrapper = styled.div`
  div {
    margin-top: 0px !important;
  }
`;

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
      rowkey: 'Property',
      value: secretDetails.property,
    },
    {
      rowkey: 'Version',
      value: secretDetails.version,
    },
  ];

  return (
    <SubRouterTabs>
      <RouterTab name="Details" path={`/details`}>
        <DetailsHeadersWrapper>
          {generateRowHeaders(secretDetailsHeaders)}
        </DetailsHeadersWrapper>
      </RouterTab>
      <RouterTab name="Events" path={`/events`}>
        <ListEventsWrapper>
          <ListEvents
            involvedObject={{
              name: externalSecretName,
              namespace: namespace || '',
              kind: 'ExternalSecret',
            }}
            clusterName={clusterName}
          />
        </ListEventsWrapper>
      </RouterTab>
      <RouterTab name="Yaml" path={`/yaml`}>
        <YamlView
          yaml={
            secretDetails?.yaml &&
            YAML.stringify(JSON.parse(secretDetails?.yaml as string))
          }
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
