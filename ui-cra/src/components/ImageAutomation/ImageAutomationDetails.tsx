import { Box } from '@material-ui/core';
import {
  EventsTable,
  Flex,
  InfoList,
  Kind,
  KubeStatusIndicator,
  RouterTab,
  SubRouterTabs,
  YamlView,
} from '@weaveworks/weave-gitops';
import { InfoField } from '@weaveworks/weave-gitops/ui/components/InfoList';
import { Header4 } from '../ProgressiveDelivery/CanaryStyles';

interface Props {
  data: any;
  kind: Kind;
  rootPath: string;
  infoFields: InfoField[];
  children?: any;
}

const ImageAutomationDetails = ({
  data,
  kind,
  rootPath,
  infoFields,
  children,
}: Props) => {
  const { name, namespace, clusterName, suspended, conditions } = data;
  return (
    <Flex wide tall column>
      <Header4>{name}</Header4>
      <Box margin={2}>
        <KubeStatusIndicator
          short
          conditions={conditions || []}
          suspended={suspended}
        />
      </Box>
      <SubRouterTabs rootPath={`${rootPath}/details`}>
        <RouterTab name="Details" path={`${rootPath}/details`}>
          <>
            <InfoList items={infoFields} />
            <Box margin={2}>{children}</Box>
          </>
        </RouterTab>
        <RouterTab name="Events" path={`${rootPath}/events`}>
          <EventsTable
            namespace={namespace}
            involvedObject={{
              kind: kind,
              name: name,
              namespace: namespace,
              clusterName: clusterName,
            }}
          />
        </RouterTab>
        <RouterTab name="yaml" path={`${rootPath}/yaml`}>
          <YamlView
            yaml={data.yaml}
            object={{
              kind: kind,
              name: name,
              namespace: namespace,
            }}
          />
        </RouterTab>
      </SubRouterTabs>
    </Flex>
  );
};

export default ImageAutomationDetails;
