import { request } from '../../utils/request';

export class PolicyService {
  static policiesUrl = '/v1/policies';

  static listPolicies = (payload: any) => {
    if (payload.limit === 20) {
      return new Promise((resolve, reject) => {
        // reject({
        //   error: 'Invalid limit',
        // });
        resolve({
          policies: [
            {
              name: 'Container Running As Root',
              id: 'weave.policies.container-running-as-rootss',
              code: 'package weave.advisor.podSecurity.runningAsRoot\n\nexclude_namespace := input.parameters.exclude_namespace\nexclude_label_key := input.parameters.exclude_label_key\nexclude_label_value := input.parameters.exclude_label_value\n\n# Check for missing securityContext.runAsNonRoot (missing in both, pod and container)\nviolation[result] {\n\tnot exclude_namespace == controller_input.metadata.namespace\n\tnot exclude_label_value == controller_input.metadata.labels[exclude_label_key]\n\n\tcontroller_spec.securityContext\n\tnot controller_spec.securityContext.runAsNonRoot\n\tnot controller_spec.securityContext.runAsNonRoot == false\n\n\tsome i\n\tcontainers := controller_spec.containers[i]\n\tcontainers.securityContext\n\tnot containers.securityContext.runAsNonRoot\n\tnot containers.securityContext.runAsNonRoot == false\n\n\tresult = {\n\t\t"issue detected": true,\n\t\t"msg": sprintf("Container missing spec.template.spec.containers[%v].securityContext.runAsNonRoot while Pod spec.template.spec.securityContext.runAsNonRoot is not defined as well.", [i]),\n\t\t"violating_key": sprintf("spec.template.spec.containers[%v].securityContext", [i]),\n\t}\n}\n\n# Container security context\n# Check if containers.securityContext.runAsNonRoot exists and = false\nviolation[result] {\n\tnot exclude_namespace == controller_input.metadata.namespace\n\tnot exclude_label_value == controller_input.metadata.labels[exclude_label_key]\n\n\tsome i\n\tcontainers := controller_spec.containers[i]\n\tcontainers.securityContext\n\tcontainers.securityContext.runAsNonRoot == false\n\n\tresult = {\n\t\t"issue detected": true,\n\t\t"msg": sprintf("Container spec.template.spec.containers[%v].securityContext.runAsNonRoot should be set to true ", [i]),\n\t\t"violating_key": sprintf("spec.template.spec.containers[%v].securityContext.runAsNonRoot", [i]),\n\t\t"recommended_value": true,\n\t}\n}\n\n# Pod security context\n# Check if spec.securityContext.runAsNonRoot exists and = false\nviolation[result] {\n\tnot exclude_namespace == controller_input.metadata.namespace\n\tnot exclude_label_value == controller_input.metadata.labels[exclude_label_key]\n\n\tcontroller_spec.securityContext\n\tcontroller_spec.securityContext.runAsNonRoot == false\n\n\tresult = {\n\t\t"issue detected": true,\n\t\t"msg": "Pod spec.template.spec.securityContext.runAsNonRoot should be set to true",\n\t\t"violating_key": "spec.template.spec.securityContext.runAsNonRoot",\n\t\t"recommended_value": true,\n\t}\n}\n\ncontroller_input = input.review.object\n\ncontroller_spec = controller_input.spec.template.spec {\n\tcontains(controller_input.kind, {"StatefulSet", "DaemonSet", "Deployment", "Job", "ReplicaSet"})\n} else = controller_input.spec {\n\tcontroller_input.kind == "Pod"\n} else = controller_input.spec.jobTemplate.spec.template.spec {\n\tcontroller_input.kind == "CronJob"\n}\n\ncontains(kind, kinds) {\n\tkinds[_] = kind\n}',
              description:
                'Running as root gives the container full access to all resources in the VM it is running on. Containers should not run with such access rights unless required by design. This Policy enforces that the `securityContext.runAsNonRoot` attribute is set to `true`. \n',
              howToSolve:
                'You should set `securityContext.runAsNonRoot` to `true`. Not setting it will default to giving the container root user rights on the VM that it is running on. \n```\n...\n  spec:\n    securityContext:\n      runAsNonRoot: true\n```\nhttps://kubernetes.io/docs/tasks/configure-pod-container/security-context/\n',
              category: 'weave.categories.pod-security',
              tags: [
                'pci-dss',
                'cis-benchmark',
                'mitre-attack',
                'nist800-190',
                'gdpr',
                'default',
              ],
              severity: 'high',
              controls: [
                'weave.controls.pci-dss.2.2.4',
                'weave.controls.pci-dss.2.2.5',
                'weave.controls.cis-benchmark.5.2.6',
                'weave.controls.mitre-attack.4.1',
                'weave.controls.nist-800-190.3.3.1',
                'weave.controls.gdpr.24',
                'weave.controls.gdpr.25',
                'weave.controls.gdpr.32',
              ],
              gitCommit: '',
              parameters: [
                {
                  name: 'exclude_namespace',
                  type: 'string',
                  value: {
                    '@type': 'type.googleapis.com/google.protobuf.StringValue',
                    value: '"kube-system"',
                  },
                  required: false,
                },
                {
                  name: 'exclude_label_key',
                  type: 'string',
                  value: null,
                  required: false,
                },
                {
                  name: 'exclude_label_value',
                  type: 'string',
                  value: null,
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
              createdAt: '2022-04-11T17:35:04+02:00',
            },
            {
              name: 'Containers Read Only Root Filesystem',
              id: 'weave.policies.containers-read-only-root-filesystemss',
              code: 'package weave.advisor.podSecurity.enforce_ro_fs\n\nread_only = input.parameters.read_only\nexclude_namespace := input.parameters.exclude_namespace\nexclude_label_key := input.parameters.exclude_label_key\nexclude_label_value := input.parameters.exclude_label_value\n\nviolation[result] {\n  not exclude_namespace == controller_input.metadata.namespace\n  not exclude_label_value == controller_input.metadata.labels[exclude_label_key]\n  some i\n  containers := controller_spec.containers[i]\n  root_fs := containers.securityContext.readOnlyRootFilesystem\n  not root_fs == read_only\n  result = {\n    "issue detected": true,\n    "msg": sprintf("readOnlyRootFilesystem should equal \'%v\'; detected \'%v\'", [read_only, root_fs]),\n    "recommended_value": read_only,\n    "violating_key": sprintf("spec.template.spec.containers[%v].securityContext.readOnlyRootFilesystem", [i]) \n  }\n}\n\n# Controller input\ncontroller_input = input.review.object\n\n# controller_container acts as an iterator to get containers from the template\ncontroller_spec = controller_input.spec.template.spec {\n  contains_kind(controller_input.kind, {"StatefulSet" , "DaemonSet", "Deployment", "Job"})\n} else = controller_input.spec {\n  controller_input.kind == "Pod"\n} else = controller_input.spec.jobTemplate.spec.template.spec {\n  controller_input.kind == "CronJob"\n}\n\ncontains_kind(kind, kinds) {\n  kinds[_] = kind\n}',
              description:
                'This Policy will cause a violation if the root file system is not mounted as specified. As a security practice, the root file system should be read-only or expose risk to your nodes if compromised. \n\nThis Policy requires containers must run with a read-only root filesystem (i.e. no writable layer).\n',
              howToSolve:
                'Set `readOnlyRootFilesystem` in your `securityContext` to the value specified in the Policy. \n```\n...\n  spec:\n    containers:\n      - securityContext:\n          readOnlyRootFilesystem: <read_only>\n```\n\nhttps://kubernetes.io/docs/concepts/policy/pod-security-policy/#volumes-and-file-systems\n',
              category: 'weave.categories.pod-security',
              tags: ['mitre-attack', 'nist800-190'],
              severity: 'high',
              controls: [
                'weave.controls.mitre-attack.3.2',
                'weave.controls.nist-800-190.4.4.4',
              ],
              gitCommit: '',
              parameters: [
                {
                  name: 'read_only',
                  type: 'boolean',
                  value: {
                    '@type': 'type.googleapis.com/google.protobuf.BoolValue',
                    value: true,
                  },
                  required: true,
                },
                {
                  name: 'exclude_namespace',
                  type: 'string',
                  value: null,
                  required: false,
                },
                {
                  name: 'exclude_label_key',
                  type: 'string',
                  value: null,
                  required: false,
                },
                {
                  name: 'exclude_label_value',
                  type: 'string',
                  value: null,
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
              createdAt: '2022-04-11T17:35:04+02:00',
            },
            {
              name: 'Container Running As Root',
              id: 'weave.policies.container-running-as-root',
              code: 'package weave.advisor.podSecurity.runningAsRoot\n\nexclude_namespace := input.parameters.exclude_namespace\nexclude_label_key := input.parameters.exclude_label_key\nexclude_label_value := input.parameters.exclude_label_value\n\n# Check for missing securityContext.runAsNonRoot (missing in both, pod and container)\nviolation[result] {\n\tnot exclude_namespace == controller_input.metadata.namespace\n\tnot exclude_label_value == controller_input.metadata.labels[exclude_label_key]\n\n\tcontroller_spec.securityContext\n\tnot controller_spec.securityContext.runAsNonRoot\n\tnot controller_spec.securityContext.runAsNonRoot == false\n\n\tsome i\n\tcontainers := controller_spec.containers[i]\n\tcontainers.securityContext\n\tnot containers.securityContext.runAsNonRoot\n\tnot containers.securityContext.runAsNonRoot == false\n\n\tresult = {\n\t\t"issue detected": true,\n\t\t"msg": sprintf("Container missing spec.template.spec.containers[%v].securityContext.runAsNonRoot while Pod spec.template.spec.securityContext.runAsNonRoot is not defined as well.", [i]),\n\t\t"violating_key": sprintf("spec.template.spec.containers[%v].securityContext", [i]),\n\t}\n}\n\n# Container security context\n# Check if containers.securityContext.runAsNonRoot exists and = false\nviolation[result] {\n\tnot exclude_namespace == controller_input.metadata.namespace\n\tnot exclude_label_value == controller_input.metadata.labels[exclude_label_key]\n\n\tsome i\n\tcontainers := controller_spec.containers[i]\n\tcontainers.securityContext\n\tcontainers.securityContext.runAsNonRoot == false\n\n\tresult = {\n\t\t"issue detected": true,\n\t\t"msg": sprintf("Container spec.template.spec.containers[%v].securityContext.runAsNonRoot should be set to true ", [i]),\n\t\t"violating_key": sprintf("spec.template.spec.containers[%v].securityContext.runAsNonRoot", [i]),\n\t\t"recommended_value": true,\n\t}\n}\n\n# Pod security context\n# Check if spec.securityContext.runAsNonRoot exists and = false\nviolation[result] {\n\tnot exclude_namespace == controller_input.metadata.namespace\n\tnot exclude_label_value == controller_input.metadata.labels[exclude_label_key]\n\n\tcontroller_spec.securityContext\n\tcontroller_spec.securityContext.runAsNonRoot == false\n\n\tresult = {\n\t\t"issue detected": true,\n\t\t"msg": "Pod spec.template.spec.securityContext.runAsNonRoot should be set to true",\n\t\t"violating_key": "spec.template.spec.securityContext.runAsNonRoot",\n\t\t"recommended_value": true,\n\t}\n}\n\ncontroller_input = input.review.object\n\ncontroller_spec = controller_input.spec.template.spec {\n\tcontains(controller_input.kind, {"StatefulSet", "DaemonSet", "Deployment", "Job", "ReplicaSet"})\n} else = controller_input.spec {\n\tcontroller_input.kind == "Pod"\n} else = controller_input.spec.jobTemplate.spec.template.spec {\n\tcontroller_input.kind == "CronJob"\n}\n\ncontains(kind, kinds) {\n\tkinds[_] = kind\n}',
              description:
                'Running as root gives the container full access to all resources in the VM it is running on. Containers should not run with such access rights unless required by design. This Policy enforces that the `securityContext.runAsNonRoot` attribute is set to `true`. \n',
              howToSolve:
                'You should set `securityContext.runAsNonRoot` to `true`. Not setting it will default to giving the container root user rights on the VM that it is running on. \n```\n...\n  spec:\n    securityContext:\n      runAsNonRoot: true\n```\nhttps://kubernetes.io/docs/tasks/configure-pod-container/security-context/\n',
              category: 'weave.categories.pod-security',
              tags: [
                'pci-dss',
                'cis-benchmark',
                'mitre-attack',
                'nist800-190',
                'gdpr',
                'default',
              ],
              severity: 'high',
              controls: [
                'weave.controls.pci-dss.2.2.4',
                'weave.controls.pci-dss.2.2.5',
                'weave.controls.cis-benchmark.5.2.6',
                'weave.controls.mitre-attack.4.1',
                'weave.controls.nist-800-190.3.3.1',
                'weave.controls.gdpr.24',
                'weave.controls.gdpr.25',
                'weave.controls.gdpr.32',
              ],
              gitCommit: '',
              parameters: [
                {
                  name: 'exclude_namespace',
                  type: 'string',
                  value: {
                    '@type': 'type.googleapis.com/google.protobuf.StringValue',
                    value: '"kube-system"',
                  },
                  required: false,
                },
                {
                  name: 'exclude_label_key',
                  type: 'string',
                  value: null,
                  required: false,
                },
                {
                  name: 'exclude_label_value',
                  type: 'string',
                  value: null,
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
              createdAt: '2022-04-11T17:35:04+02:00',
            },
            {
              name: 'Containers Read Only Root Filesystem',
              id: 'weave.policies.containers-read-only-root-filesystem',
              code: 'package weave.advisor.podSecurity.enforce_ro_fs\n\nread_only = input.parameters.read_only\nexclude_namespace := input.parameters.exclude_namespace\nexclude_label_key := input.parameters.exclude_label_key\nexclude_label_value := input.parameters.exclude_label_value\n\nviolation[result] {\n  not exclude_namespace == controller_input.metadata.namespace\n  not exclude_label_value == controller_input.metadata.labels[exclude_label_key]\n  some i\n  containers := controller_spec.containers[i]\n  root_fs := containers.securityContext.readOnlyRootFilesystem\n  not root_fs == read_only\n  result = {\n    "issue detected": true,\n    "msg": sprintf("readOnlyRootFilesystem should equal \'%v\'; detected \'%v\'", [read_only, root_fs]),\n    "recommended_value": read_only,\n    "violating_key": sprintf("spec.template.spec.containers[%v].securityContext.readOnlyRootFilesystem", [i]) \n  }\n}\n\n# Controller input\ncontroller_input = input.review.object\n\n# controller_container acts as an iterator to get containers from the template\ncontroller_spec = controller_input.spec.template.spec {\n  contains_kind(controller_input.kind, {"StatefulSet" , "DaemonSet", "Deployment", "Job"})\n} else = controller_input.spec {\n  controller_input.kind == "Pod"\n} else = controller_input.spec.jobTemplate.spec.template.spec {\n  controller_input.kind == "CronJob"\n}\n\ncontains_kind(kind, kinds) {\n  kinds[_] = kind\n}',
              description:
                'This Policy will cause a violation if the root file system is not mounted as specified. As a security practice, the root file system should be read-only or expose risk to your nodes if compromised. \n\nThis Policy requires containers must run with a read-only root filesystem (i.e. no writable layer).\n',
              howToSolve:
                'Set `readOnlyRootFilesystem` in your `securityContext` to the value specified in the Policy. \n```\n...\n  spec:\n    containers:\n      - securityContext:\n          readOnlyRootFilesystem: <read_only>\n```\n\nhttps://kubernetes.io/docs/concepts/policy/pod-security-policy/#volumes-and-file-systems\n',
              category: 'weave.categories.pod-security',
              tags: ['mitre-attack', 'nist800-190'],
              severity: 'high',
              controls: [
                'weave.controls.mitre-attack.3.2',
                'weave.controls.nist-800-190.4.4.4',
              ],
              gitCommit: '',
              parameters: [
                {
                  name: 'read_only',
                  type: 'boolean',
                  value: {
                    '@type': 'type.googleapis.com/google.protobuf.BoolValue',
                    value: true,
                  },
                  required: true,
                },
                {
                  name: 'exclude_namespace',
                  type: 'string',
                  value: null,
                  required: false,
                },
                {
                  name: 'exclude_label_key',
                  type: 'string',
                  value: null,
                  required: false,
                },
                {
                  name: 'exclude_label_value',
                  type: 'string',
                  value: null,
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
              createdAt: '2022-04-11T17:35:04+02:00',
            },
            {
              name: 'Containers Running With Privilege Escalation',
              id: 'weave.policies.containers-running-with-privilege-escalation',
              code: 'package weave.advisor.podSecurity.privilegeEscalation\n\nexclude_namespace := input.parameters.exclude_namespace\nallow_privilege_escalation := input.parameters.allow_privilege_escalation\nexclude_label_key := input.parameters.exclude_label_key\nexclude_label_value := input.parameters.exclude_label_value\n\nviolation[result] {\n  some i\n  isExcludedNamespace == false\n  not exclude_namespace == controller_input.metadata.namespace\n  not exclude_label_value == controller_input.metadata.labels[exclude_label_key]\n  containers := controller_spec.containers[i]\n  allow_priv := containers.securityContext.allowPrivilegeEscalation\n  not allow_priv == allow_privilege_escalation\n  result = {\n    "issue detected": true,\n    "msg": sprintf("Container\'s privilegeEscalation should be set to \'%v\'; detected \'%v\'", [allow_privilege_escalation, allow_priv]),\n    "violating_key": sprintf("spec.template.spec.containers[%v].securityContext.allowPrivilegeEscalation", [i]),\n    "recommended_value": allow_privilege_escalation\n  }\n}\n\nisExcludedNamespace  = true {\n  input.review.object.metadata.namespace == exclude_namespace\n}else = false {true}\n\nis_array_contains(array,str) {\n  array[_] = str\n}\n\n# Controller input\ncontroller_input = input.review.object\n\n# controller_container acts as an iterator to get containers from the template\ncontroller_spec = controller_input.spec.template.spec {\n  contains_kind(controller_input.kind, {"StatefulSet" , "DaemonSet", "Deployment", "Job"})\n} else = controller_input.spec {\n  controller_input.kind == "Pod"\n} else = controller_input.spec.jobTemplate.spec.template.spec {\n  controller_input.kind == "CronJob"\n}\n\ncontains_kind(kind, kinds) {\n  kinds[_] = kind\n}',
              description:
                'Containers are running with PrivilegeEscalation configured. Setting this Policy to `true` allows child processes to gain more privileges than its parent process.  \n\nThis Policy gates whether or not a user is allowed to set the security context of a container to `allowPrivilegeEscalation` to `true`. The default value for this is `false` so no child process of a container can gain more privileges than its parent.\n\nThere are 2 parameters for this Policy:\n- exclude_namespace (string) : This sets a namespace you want to exclude from Policy compliance checking. \n- allow_privilege_escalation (bool) : This checks for the value of `allowPrivilegeEscalation` in your spec.  \n',
              howToSolve:
                'Check the following path to see what the PrivilegeEscalation value is set to.\n```\n...\n  spec:\n    containers:\n      securityContext:\n        allowPrivilegeEscalation: <value>\n```\nhttps://kubernetes.io/docs/tasks/configure-pod-container/security-context/\n',
              category: 'weave.categories.pod-security',
              tags: [
                'pci-dss',
                'cis-benchmark',
                'mitre-attack',
                'nist800-190',
                'gdpr',
                'default',
                'soc2-type1',
              ],
              severity: 'high',
              controls: [
                'weave.controls.pci-dss.2.2.4',
                'weave.controls.pci-dss.2.2.5',
                'weave.controls.cis-benchmark.5.2.5',
                'weave.controls.mitre-attack.4.1',
                'weave.controls.nist-800-190.3.3.2',
                'weave.controls.gdpr.25',
                'weave.controls.gdpr.32',
                'weave.controls.gdpr.24',
                'weave.controls.soc2-type-i.1.6.1',
              ],
              gitCommit: '',
              parameters: [
                {
                  name: 'exclude_namespace',
                  type: 'string',
                  value: {
                    '@type': 'type.googleapis.com/google.protobuf.StringValue',
                    value: '"kube-system"',
                  },
                  required: true,
                },
                {
                  name: 'allow_privilege_escalation',
                  type: 'boolean',
                  value: {
                    '@type': 'type.googleapis.com/google.protobuf.BoolValue',
                    value: false,
                  },
                  required: true,
                },
                {
                  name: 'exclude_label_key',
                  type: 'string',
                  value: null,
                  required: false,
                },
                {
                  name: 'exclude_label_value',
                  type: 'string',
                  value: null,
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
              createdAt: '2022-04-11T17:35:04+02:00',
            },
          ],
          total: 5,
        });
      });
    }
    return request('GET', this.policiesUrl, {
      cache: 'no-store',
    });
  };
  static getPolicyById = (id: string) => {
    return request('GET', `${this.policiesUrl}/${id}`, {
      cache: 'no-store',
    });
  };

  // TODO payload should be a ClusterId
  static listPolicyViolations = () => {
    return request('POST', `${this.policiesUrl}/violations`, {
      cache: 'no-store',
    });
    // return new Promise<ListPolicyValidationsResponse>((resolve, reject) => {
    //   resolve({
    //     violations: [
    //       {
    //         id: '1',
    //         severity: 'high',
    //         category: 'weave.categories.pod-security',
    //         message: 'Pod Security Policy violation',
    //         namespace: 'default',
    //         entity: 'default/weave-net-0',
    //       },
    //       {
    //         id: '2',
    //         severity: 'high',
    //         category: 'weave.categories.pod-security',
    //         message: 'Pod Security Policy violation',
    //         namespace: 'default',
    //         entity: 'default/weave-net-1',
    //       },
    //       {
    //         id: '3',
    //         severity: 'high',
    //         category: 'weave.categories.pod-security',
    //         message: 'Pod Security Policy violation',
    //         namespace: 'default',
    //         entity: 'default/weave-net-2',
    //       },
    //     ],
    //     total: 3,
    //   });
    // });
  };
}
