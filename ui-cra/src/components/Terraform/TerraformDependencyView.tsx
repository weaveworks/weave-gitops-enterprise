import { DagGraph, Flex, MessageBox } from '@weaveworks/weave-gitops';
import { FluxObjectNode } from '@weaveworks/weave-gitops/ui/lib/objects';
import React from 'react';
import styled from 'styled-components';
import { TerraformObject } from '../../api/terraform/types.pb';
import { useListTerraformObjects } from '../../contexts/Terraform';
import { Body, Title } from '../Shared';
import {
  getGraphNodes,
  makeObjectId,
  TerraformNodesMap,
  TerraformObjectNode,
} from './dependencies';

type Props = {
  className?: string;
  object: TerraformObject;
};

function TerraformDependenciesView({ className, object }: Props) {
  const { isLoading, data, error } = useListTerraformObjects();
  const [graphNodes, setGraphNodes] = React.useState<TerraformObjectNode[]>([]);

  React.useEffect(() => {
    if (isLoading) {
      return;
    }

    if (error) {
      setGraphNodes([]);
      return;
    }

    const allNodes: TerraformNodesMap = {};
    data?.objects?.forEach(obj => {
      const id = makeObjectId(obj.namespace, obj.name);
      allNodes[id] = {
        ...obj,
        type: 'Terraform',
        id: id,
        isCurrentNode: false,
        parentIds:
          obj?.dependsOn?.map(dependency => {
            const namespace = dependency.namespace || obj.namespace;

            return namespace + '/' + dependency.name;
          }) || [],
      };
    });

    const nodes = getGraphNodes(allNodes, object);

    nodes.sort((a, b) => a.id.localeCompare(b.id));

    if (nodes.length === 0) {
      setGraphNodes([]);
    } else {
      setGraphNodes(nodes);
    }
  }, [isLoading, data, error, object]);

  const shouldShowGraph = graphNodes && graphNodes.length;

  return (
    <Flex align wide tall column>
      {shouldShowGraph ? (
        <DagGraph
          className={className}
          nodes={graphNodes as FluxObjectNode[]}
        />
      ) : (
        <MessageBox>
          <Title>No Dependencies</Title>
          <Body>
            There are no dependencies set up for your Terraform object at this
            time. You can set them up using the "dependsOn" field.
          </Body>
          <Title>What are dependencies for?</Title>
          <Body>
            Dependencies allow you to relate different Terraform objects, as
            well as specifying an order in which your resources should be
            started. For example, you can wait for a database to report as
            'Ready' before attempting to deploy other services.
          </Body>
        </MessageBox>
      )}
    </Flex>
  );
}

export default styled(TerraformDependenciesView).attrs({
  className: TerraformDependenciesView.name,
})`
  width: 100%;
`;
