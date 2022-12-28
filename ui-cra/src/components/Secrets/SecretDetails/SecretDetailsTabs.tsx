import { RouterTab } from '@weaveworks/weave-gitops';
import { GetExternalSecretResponse } from '../../../cluster-services/cluster_services.pb';
import styled from 'styled-components';
import { generateRowHeaders, SectionRowHeader } from '../../RowHeader';
import { Routes } from '../../../utils/nav';
import { CustomSubRouterTabs } from '../../Workspaces/WorkspaceStyles';
import ListEvents from '../../ProgressiveDelivery/CanaryDetails/Events/ListEvents';

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
    <div>
      <CustomSubRouterTabs rootPath={`${path}/secretDetails`}>
        <RouterTab name="Details" path={`${path}/secretDetails`}>
          <DetailsHeadersWrapper>
            {generateRowHeaders(secretDetailsHeaders)}
          </DetailsHeadersWrapper>
        </RouterTab>

        <RouterTab name="Events" path={`${path}/events`}>
          <ListEventsWrapper>
            <ListEvents
              clusterName={clusterName}
              involvedObject={{
                kind: secretDetails?.secretName,
                name: externalSecretName,
                namespace: namespace || '',
              }}
            />
          </ListEventsWrapper>
        </RouterTab>
      </CustomSubRouterTabs>
    </div>
  );
};

export default SecretDetailsTabs;
