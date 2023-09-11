import { TerraformObject } from '../../api/terraform/types.pb';

export interface TerraformObjectNode extends TerraformObject {
  id: string;
  isCurrentNode: boolean;
  type: string;
  parentIds: string[];
}

export type TerraformNodesMap = { [key: string]: TerraformObjectNode };

export function makeObjectId(namespace?: string, name?: string) {
  return namespace + '/' + name;
}

export function getNeighborNodes(
  nodes: TerraformNodesMap,
  currentNode: TerraformObject,
): TerraformObjectNode[] {
  let dependencyNodes: TerraformObjectNode[] = [];
  if (currentNode.dependsOn) {
    dependencyNodes = currentNode.dependsOn
      .map(dependency => {
        const name = dependency.name;
        let namespace = dependency.namespace;

        if (!namespace) {
          namespace = currentNode.namespace;
        }

        return nodes[makeObjectId(namespace, name)];
      })
      .filter(n => n);
  }

  const nodesArray: TerraformObjectNode[] = Object.values(nodes);

  const dependentNodes = nodesArray.filter(node => {
    let isDependent = false;

    for (const dependency of node.dependsOn || []) {
      const name = dependency.name;
      let namespace = dependency.namespace;
      if (!namespace) {
        namespace = node.namespace;
      }

      if (name === currentNode.name && namespace === currentNode.namespace) {
        isDependent = true;
        break;
      }
    }

    return isDependent;
  });

  return dependencyNodes.concat(dependentNodes);
}

// getGraphNodes returns all nodes in the current node's dependency tree, including the current node.
export function getGraphNodes(
  nodes: TerraformNodesMap,
  object: TerraformObject,
): TerraformObjectNode[] {
  // Find node, corresponding to the object.
  const currentNode = nodes[makeObjectId(object.namespace, object.name)];

  if (!currentNode) {
    return [];
  }

  currentNode.isCurrentNode = true;

  // Find nodes in the current node's dependency tree.
  let graphNodes: TerraformObjectNode[] = [];

  const visitedNodes: { [name: string]: boolean } = {};
  visitedNodes[currentNode.id] = true;
  let nodesToExplore: TerraformObjectNode[] = [currentNode];

  while (nodesToExplore.length > 0) {
    const node = nodesToExplore.shift();

    const newNodes = getNeighborNodes(nodes, node || {}).filter(
      n => !visitedNodes[n.id],
    );

    for (const n of newNodes) {
      visitedNodes[n.id] = true;
    }

    nodesToExplore = nodesToExplore.concat(newNodes);

    graphNodes = graphNodes.concat(node || []);
  }

  if (graphNodes.length === 1) {
    return [];
  }

  return graphNodes;
}
