import { request } from '../../utils/request';

export class PolicyService {
  static policiesUrl = '/v1/policies';

  static getPolicyList = (payload: any) => {
    console.log(payload);

    if (payload.limit === 20) {
      return new Promise((resolve, reject) => {
        resolve({
          policies: [
            {
              name: 'MariaDB Enforce Environment Variable - MYSQL_ROOT_PASSWORD',
              id: 'magalix.policies.mariadb-enforce-mysql-root-password-env-var',
              code: 'package magalix.advisor.mariadb.enforce_mysql_root_password_env_var\n\nenv_name = "MYSQL_ROOT_PASSWORD"\napp_name = "mariadb"\nexclude_namespace = input.parameters.exclude_namespace\nexclude_label_key := input.parameters.exclude_label_key\nexclude_label_value := input.parameters.exclude_label_value\n\n\nviolation[result] {\n  not exclude_namespace == controller_input.metadata.namespace\n  not exclude_label_value == controller_input.metadata.labels[exclude_label_key]\n  some i\n  containers := controller_spec.containers[i]\n  contains(containers.image, app_name)\n  not containers.env\n  result = {\n    "issue detected": true,\n    "msg": "environment variables needs to be set",\n    "violating_key": sprintf("spec.template.spec.containers[%v]", [i]),  }\n}\n\nviolation[result] {\n  not exclude_namespace == controller_input.metadata.namespace\n  not exclude_label_value == controller_input.metadata.labels[exclude_label_key]\n  some i\n  containers := controller_spec.containers[i]\n  contains(containers.image, app_name)\n  envs := containers.env\n  not array_contains(envs, env_name)\n  result = {\n    "issue detected": true,\n    "msg": sprintf("\'%v\' is missing\'; detected \'%v\'", [env_name, envs]),\n    "violating_key": sprintf("spec.template.spec.containers[%v].env.name", [i])\n  }\n}\n\n\narray_contains(array, element) {\n  array[_].name = element\n}\n\n# Controller input\ncontroller_input = input.review.object\n\n# controller_container acts as an iterator to get containers from the template\ncontroller_spec = controller_input.spec.template.spec {\n  contains_kind(controller_input.kind, {"StatefulSet" , "DaemonSet", "Deployment", "Job"})\n} else = controller_input.spec {\n  controller_input.kind == "Pod"\n} else = controller_input.spec.jobTemplate.spec.template.spec {\n  controller_input.kind == "CronJob"\n}\n\ncontains_kind(kind, kinds) {\n  kinds[_] = kind\n}',
              description:
                'This Policy ensures MYSQL_ROOT_PASSWORD environment variable are in place when using the official container images from Docker Hub.\nMYSQL_ROOT_PASSWORD: The MYSQL_ROOT_PASSWORD environment variable specifies a password for the MARIADB root account. \n',
              howToSolve:
                'If you encounter a violation, ensure the MYSQL_ROOT_PASSWORD environment variables is set.\nFor futher information about the MariaDB Docker container, check here: https://hub.docker.com/_/mariadb\n',
              category: 'magalix.categories.access-control',
              tags: ['pci-dss', 'mitre-attack', 'hipaa', 'gdpr'],
              severity: 'high',
              controls: [
                'magalix.controls.pci-dss.2.1',
                'magalix.controls.hipaa.164.312.a.1',
                'magalix.controls.hipaa.164.312.a.2.i',
                'magalix.controls.gdpr.24',
                'magalix.controls.gdpr.25',
                'magalix.controls.gdpr.32',
              ],
              gitCommit: '',
              parameters: [
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
                  type: 'string',
                  default: null,
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
            {
              name: 'Missing Owner Label',
              id: 'magalix.policies.missing-owner-label',
              code: 'package magalix.advisor.labels.missing_owner_label\n\nexclude_namespace := input.parameters.exclude_namespace\nexclude_label_key := input.parameters.exclude_label_key\nexclude_label_value := input.parameters.exclude_label_value\n\nviolation[result] {\n  not exclude_namespace == controller_input.metadata.namespace\n  not exclude_label_value == controller_input.metadata.labels[exclude_label_key]\n  # Filter the type of entity before moving on since this shouldn\'t apply to all entities\n  label := "owner"\n  contains_kind(controller_input.kind, {"StatefulSet" , "DaemonSet", "Deployment", "Job"})\n  not controller_input.metadata.labels[label]\n  result = {\n    "issue detected": true,\n    "msg": sprintf("you are missing a label with the key \'%v\'", [label]),\n    "violating_key": "metadata.labels",\n    "recommended_value": label\n  }\n}\n\n# Controller input\ncontroller_input = input.review.object\n\ncontains_kind(kind, kinds) {\n  kinds[_] = kind\n}',
              description:
                "Custom labels can help enforce organizational standards for each artifact deployed. This Policy ensure a custom label key is set in the entity's `metadata`. The Policy detects the presence of the following: \n\n### owner\nA label key of `owner` will help identify who the owner of this entity is. \n\n### app.kubernetes.io/name\nThe name of the application\t\n\n### app.kubernetes.io/instance\nA unique name identifying the instance of an application\t  \n\n### app.kubernetes.io/version\nThe current version of the application (e.g., a semantic version, revision hash, etc.)\n\n### app.kubernetes.io/part-of\nThe name of a higher level application this one is part of\t\n\n### app.kubernetes.io/managed-by\nThe tool being used to manage the operation of an application\t\n\n### app.kubernetes.io/created-by\nThe controller/user who created this resource\t\n",
              howToSolve:
                'Add these custom labels to `metadata`.\n* owner\n* app.kubernetes.io/name\n* app.kubernetes.io/instance\n* app.kubernetes.io/version\n* app.kubernetes.io/name\n* app.kubernetes.io/part-of\n* app.kubernetes.io/managed-by\n* app.kubernetes.io/created-by\n\n```\nmetadata:\n  labels:\n    <label>: value\n```  \nFor additional information, please check\n* https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels \n',
              category: 'magalix.categories.organizational-standards',
              tags: [],
              severity: 'low',
              controls: [],
              gitCommit: '',
              parameters: [
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
                  type: 'string',
                  default: null,
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
            {
              name: 'Mongo-Express Enforce Environment Variable - ME_CONFIG_BASICAUTH_PASSWORD',
              id: 'magalix.policies.mongo-express-enforce-auth-password-env-var',
              code: 'package magalix.advisor.mongo_express.enforce_auth_password_env_var\n\nenv_name = "ME_CONFIG_BASICAUTH_PASSWORD"\napp_name = "mongo-express"\nexclude_namespace = input.parameters.exclude_namespace\nexclude_label_key := input.parameters.exclude_label_key\nexclude_label_value := input.parameters.exclude_label_value\n\n\nviolation[result] {\n  not exclude_namespace == controller_input.metadata.namespace\n  not exclude_label_value == controller_input.metadata.labels[exclude_label_key]\n  some i\n  containers := controller_spec.containers[i]\n  contains(containers.image, app_name)\n  not containers.env\n  result = {\n    "issue detected": true,\n    "msg": "environment variables needs to be set",\n    "violating_key": sprintf("spec.template.spec.containers[%v]", [i]),  }\n}\n\nviolation[result] {\n  not exclude_namespace == controller_input.metadata.namespace\n  not exclude_label_value == controller_input.metadata.labels[exclude_label_key]\n  some i\n  containers := controller_spec.containers[i]\n  contains(containers.image, app_name)\n  envs := containers.env\n  not array_contains(envs, env_name)\n  result = {\n    "issue detected": true,\n    "msg": sprintf("\'%v\' is missing\'; detected \'%v\'", [env_name, envs]),\n    "violating_key": sprintf("spec.template.spec.containers[%v].env.name", [i])\n  }\n}\n\n\narray_contains(array, element) {\n  array[_].name = element\n}\n\n# Controller input\ncontroller_input = input.review.object\n\n# controller_container acts as an iterator to get containers from the template\ncontroller_spec = controller_input.spec.template.spec {\n  contains_kind(controller_input.kind, {"StatefulSet" , "DaemonSet", "Deployment", "Job"})\n} else = controller_input.spec {\n  controller_input.kind == "Pod"\n} else = controller_input.spec.jobTemplate.spec.template.spec {\n  controller_input.kind == "CronJob"\n}\n\ncontains_kind(kind, kinds) {\n  kinds[_] = kind\n}',
              description:
                'This Policy ensures ME_CONFIG_BASICAUTH_PASSWORD environment variable are in place when using the official container images from Docker Hub.\nME_CONFIG_BASICAUTH_PASSWORD: The ME_CONFIG_BASICAUTH_PASSWORD environment variable sets the mongo-express web password.\n',
              howToSolve:
                'If you encounter a violation, ensure the ME_CONFIG_BASICAUTH_PASSWORD environment variables is set.\nFor futher information about the Mongo-Express Docker container, check here: https://hub.docker.com/_/mongo-express\n',
              category: 'magalix.categories.access-control',
              tags: ['pci-dss', 'mitre-attack', 'hipaa'],
              severity: 'high',
              controls: [
                'magalix.controls.pci-dss.2.1',
                'magalix.controls.hipaa.164.312.a.1',
                'magalix.controls.hipaa.164.312.a.2.i',
              ],
              gitCommit: '',
              parameters: [
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
                  type: 'string',
                  default: null,
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
            {
              name: 'Persistent Volume Reclaim Policy Should Be Set To Retain',
              id: 'magalix.policies.persistent-volume-reclaim-policy-should-be-set-to-retain',
              code: 'package magalix.advisor.storage.persistentvolume_reclaim_policy\n\npolicy := input.parameters.pv_policy\nexclude_label_key := input.parameters.exclude_label_key\nexclude_label_value := input.parameters.exclude_label_value\n\nviolation[result] {\n  pv_policy := storage_spec.persistentVolumeReclaimPolicy\n  not pv_policy == policy\n  result = {\n    "issue detected": true,\n    "msg": sprintf("persistentVolumeReclaimPolicy must be \'%v\'; found \'%v\'", [policy, pv_policy]),\n    "violating_key": "spec.persistentVolumeReclaimPolicy",\n    "recommened_value": policy\n  }\n}\n\n# controller_container acts as an iterator to get containers from the template\nstorage_spec = input.review.object.spec {\n  contains_kind(input.review.object.kind, {"PersistentVolume"})\n}\n\ncontains_kind(kind, kinds) {\n  kinds[_] = kind\n}\n\n',
              description:
                'This Policy checks to see whether or not the persistent volume reclaim policy is set.\n\nPersistentVolumes can have various reclaim policies, including "Retain", "Recycle", and "Delete". For dynamically provisioned PersistentVolumes, the default reclaim policy is "Delete". This means that a dynamically provisioned volume is automatically deleted when a user deletes the corresponding PersistentVolumeClaim. This automatic behavior might be inappropriate if the volume contains precious data. In that case, it is more appropriate to use the "Retain" policy. With the "Retain" policy, if a user deletes a PersistentVolumeClaim, the corresponding PersistentVolume is not be deleted. Instead, it is moved to the Released phase, where all of its data can be manually recovered.\n',
              howToSolve:
                'Check your reclaim policy configuration within your PersistentVolume configuration. \n```\nspec:\n  persistentVolumeReclaimPolicy: <pv_policy>\n```\n\nhttps://kubernetes.io/docs/tasks/administer-cluster/change-pv-reclaim-policy/#why-change-reclaim-policy-of-a-persistentvolume\n',
              category: 'magalix.categories.data-protection',
              tags: ['pci-dss', 'soc2-type1'],
              severity: 'medium',
              controls: [
                'magalix.controls.pci-dss.3.1',
                'magalix.controls.soc2-type-i.2.1.2',
              ],
              gitCommit: '',
              parameters: [
                {
                  name: 'pv_policy',
                  type: 'string',
                  default: {
                    '@type': 'type.googleapis.com/google.protobuf.StringValue',
                    value: '"Retain"',
                  },
                  required: true,
                },
                {
                  name: 'exclude_label_key',
                  type: 'string',
                  default: null,
                  required: false,
                },
                {
                  name: 'exclude_label_value',
                  type: 'string',
                  default: null,
                  required: false,
                },
              ],
              targets: {
                kinds: ['PersistentVolume'],
                labels: [],
                namespaces: [],
              },
              createdAt: '2022-03-31 11:36:02 +0000 UTC',
            },
          ],
          total: 4,
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
}
