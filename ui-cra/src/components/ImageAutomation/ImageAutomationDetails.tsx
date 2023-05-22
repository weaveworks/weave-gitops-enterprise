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
import styled from 'styled-components';

const HeaderSection = styled.div`
  font-size: ${({ theme }) => theme.fontSizes.large};
  margin-bottom: ${({ theme }) => theme.spacing.small};
  .title {
    margin-bottom: ${({ theme }) => theme.spacing.small};
    font-weight: bold;
  }
  svg {
    height: 24px;
    width: 24px;
  }
`;
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
      <HeaderSection>
        <div className="title">{name}</div>

        <KubeStatusIndicator
          conditions={conditions || []}
          suspended={suspended}
        />
      </HeaderSection>

      <SubRouterTabs>
        <RouterTab name="Details" path={`${rootPath}/details`}>
          <>
            <InfoList items={infoFields} />
            <Box marginTop={2}>{children}</Box>
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
