import {
  getGraphNodes,
  getNeighborNodes,
  makeObjectId,
  TerraformNodesMap,
  TerraformObjectNode,
} from '../dependencies';

describe('dependencies', () => {
  const sharedFields = {
    obj: {},
    uid: 'some-uid',
    type: 'Terraform',
    suspended: false,
    conditions: [
      {
        type: 'Ready',
        status: 'True',
        reason: 'ReconciliationSucceeded',
        message:
          'Applied revision: main/9e0930cfa1aafef1d8925d2c7b71272b0878aac4',
        timestamp: '2022-09-12T00:31:32Z',
      },
      {
        message: 'ReconciliationSucceeded',
        reason: 'ReconciliationSucceeded',
        status: 'True',
        type: 'Healthy',
      },
    ],
    isCurrentNode: false,
    clusterName: 'cluster',
    yaml: 'yaml',
  };

  const nodes: TerraformObjectNode[] = [
    {
      ...sharedFields,
      name: 'terraform1',
      namespace: 'default',
      dependsOn: [],
      parentIds: [],
      id: 'default/terraform1',
    },
    {
      ...sharedFields,
      name: 'terraforma',
      namespace: 'default',
      dependsOn: [],
      parentIds: [],
      id: 'default/terraforma',
    },
    {
      ...sharedFields,
      name: 'terraformb',
      namespace: 'default',
      dependsOn: [
        {
          name: 'terraforma',
        },
      ],
      parentIds: ['default/terraforma'],

      id: 'default/terraformb',
    },
    {
      ...sharedFields,
      name: 'terraformc',
      namespace: 'default',
      dependsOn: [
        {
          name: 'terraforma',
        },
      ],
      parentIds: ['default/terraforma'],
      id: 'default/terraformc',
    },
    {
      ...sharedFields,
      name: 'terraformd',
      namespace: 'default',
      dependsOn: [
        {
          name: 'terraformb',
        },
      ],
      parentIds: ['default/terraformb'],
      id: 'default/terraformd',
    },
    {
      ...sharedFields,
      name: 'terraforme',
      namespace: 'default',
      dependsOn: [
        {
          name: 'terraforma',
        },
        {
          name: 'terraform1',
        },
      ],
      parentIds: ['default/terraforma', 'default/terraform1'],
      id: 'default/terraforme',
    },
  ];

  describe('getNeighborNodes', () => {
    it('returns correct neighbor nodes', () => {
      const nodesMap: TerraformNodesMap = {};
      nodes.forEach(node => {
        nodesMap[node.id] = node;
      });

      const node = nodesMap[makeObjectId('default', 'terraformb')];

      expect(node).toBe(nodes[2]);

      const neighborNodes = getNeighborNodes(nodesMap, node);

      expect(neighborNodes.length).toEqual(2);
      expect(neighborNodes[0]).toBe(nodes[1]);
      expect(neighborNodes[1]).toBe(nodes[4]);
    });
  });
  describe('getGraphNodes', () => {
    it('returns correct graph nodes', () => {
      const terraformC: TerraformObjectNode = {
        type: 'Terraform',
        name: 'terraformc',
        namespace: 'default',
        dependsOn: [{ name: 'terraforma' }],
        appliedRevision: '6.2.0/79f81383288bf6542fcb5bdd8144b826b33b36e7',
        id: 'terraformc/default',
        isCurrentNode: false,
        parentIds: [],
      };

      const mappedNodes: TerraformNodesMap = {};
      nodes.forEach(node => {
        mappedNodes[node.id] = node;
      });

      const graphNodes = getGraphNodes(mappedNodes, terraformC);

      expect(graphNodes.length).toEqual(6);
      expect(graphNodes[0]).toBe(nodes[3]);
      expect(graphNodes[1]).toBe(nodes[1]);
      expect(graphNodes[2]).toBe(nodes[2]);
      expect(graphNodes[3]).toBe(nodes[5]);
      expect(graphNodes[4]).toBe(nodes[4]);
      expect(graphNodes[5]).toBe(nodes[0]);
    });
  });
});
