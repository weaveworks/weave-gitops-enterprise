import { request } from '../../utils/request';

export class PolicyService {
  static policiesUrl = '/v1/policies';

  static getPolicyList = () => {
    return request('GET', this.policiesUrl, {
      cache: 'no-store',
    });
  };
  static getPolicyById = (id: string) => {
    // return request('GET', `${this.policiesUrl}/${id}`, {
    //   cache: 'no-store',
    // });
    return new Promise((resolve, reject) => {
      resolve({
        policy: {
          name: 'Containers Block Ports Range',
          id: 'magalix.policies.containers-block-ports-range',
          code: 'package magalix.advisor.pods.block_ports\n\ntarget_port := input.parameters.target_port\nexclude_namespace := input.parameters.exclude_namespace\nexclude_label_key := input.parameters.exclude_label_key\nexclude_label_value := input.parameters.exclude_label_value\n\nviolation[result] {\n  not exclude_namespace == controller_input.metadata.namespace\n  not exclude_label_value == controller_input.metadata.labels[exclude_label_key]\n  some i,j\n  containers := controller_spec.containers[i]\n  container_ports := containers.ports[j]\n  port := container_ports.containerPort\n  not port >= target_port\n  result = {\n    "issue detected": true,\n    "msg": sprintf("containerPort is not greater than \'%v\'; found %v", [target_port, port]),\n    "violating_key": sprintf("spec.template.spec.containers[%v].ports[%v].containerPort", [i,j]) \n  }\n}\n\n# Controller input\ncontroller_input = input.review.object\n\n# controller_container acts as an iterator to get containers from the template\ncontroller_spec = controller_input.spec.template.spec {\n  contains_kind(controller_input.kind, {"StatefulSet" , "DaemonSet", "Deployment", "Job"})\n} else = controller_input.spec {\n  controller_input.kind == "Pod"\n} else = controller_input.spec.jobTemplate.spec.template.spec {\n  controller_input.kind == "CronJob"\n}\n\ncontains_kind(kind, kinds) {\n  kinds[_] = kind\n}',
          description:
            'This Policy checks for container ports that are set below the set value. TCP ports under 1024 are reserved so we recommend setting your Policy to 1024 or higher. \n',
          howToSolve:
            'Choose ports over the value that is specified in the Policy. \n```\n...\n  spec:\n    containers:\n      - ports:\n        - containerPort: <target_port>\n```\nhttps://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.txt\n',
          category: 'magalix.categories.network-security',
          tags: ['pci-dss', 'nist800-190'],
          severity: 'high',
          controls: [
            'magalix.controls.pci-dss.1.1.6',
            'magalix.controls.pci-dss.2.2.2',
            'magalix.controls.nist-800-190.4.4.2',
          ],
          gitCommit: '',
          parameters: [
            {
              name: 'target_port',
              type: 'integer',
              default: {
                '@type': 'type.googleapis.com/google.protobuf.Int32Value',
                value: 1024,
              },
              required: true,
            },
            {
              name: 'exclude_namespace',
              type: 'string',
              default: null,
              required: false,
            },
            {
              name: 'exclude_label_key',
              type: 'string',
              default: null,
              required: false,
            },
            {
              name: 'exclude_label_value',
              type: 'array',
              default: {
                '@type':
                  'type.googleapis.com/capi_server.v1.PolicyParamRepeatedString',
                value: ['ddfdf'],
              },
              required: false,
            },
          ],
          targets: {
            kinds: [
              'Deployment',
              'Job',
              'ReplicationController',
              'ReplicaSet',
              'DaemonSet',
              'StatefulSet',
              'CronJob',
            ],
            labels: [],
            namespaces: [],
          },
          createdAt: '2022-03-31 11:36:02 +0000 UTC',
        },
      });
    });
  };
}
