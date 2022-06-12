import { request } from '../../utils/request';

export class CanaryService {
  static canariesArr = [
    {
      name: 'backend',
      namespace: 'default',
      clusterName: 'Kind',
      provider: 'traefik',
      status: {
        phase: 'Failed',
        failedChecks: 0,
        canaryWeight: 0,
        iterations: 0,
        lastTransitionTime: '2022-05-11T13:54:51Z',
        conditions: [
          {
            type: 'Promoted',
            status: 'True',
            lastUpdateTime: '2022-05-11T13:54:51Z',
            lastTransitionTime: '2022-05-11T13:54:51Z',
            reason: 'Failed',
            message:
              'Canary analysis completed successfully, promotion finished.',
          },
        ],
      },
      targetDeployment: {
        uid: '4b871207-63e7-8974-b067-395c59b3676b',
        resource_version: '',
      },
      targetReference: {
        kind: 'Deployment',
        name: 'hello-world',
      },
    },
    {
      namespace: 'hello-world',
      name: 'hello-world',
      clusterName: 'Default',
      provider: 'traefik',
      targetReference: {
        kind: 'Deployment',
        name: 'hello-world',
      },
      targetDeployment: {
        uid: '4b871207-63e7-4981-b067-395c59b3676b',
        resourceVersion: '1997',
        fluxLabels: {
          kustomizeNamespace: 'hello-world',
          kustomizeName: 'hello-world',
        },
      },
      status: {
        phase: 'Initialized',
        lastTransitionTime: '2022-06-03T12:36:23Z',
        conditions: [
          {
            type: 'Promoted',
            status: 'True',
            lastUpdateTime: '2022-06-03T12:36:23Z',
            lastTransitionTime: '2022-06-03T12:36:23Z',
            reason: 'Initialized',
            message: 'Deployment initialization completed.',
          },
        ],
      },
    },
    {
      name: 'hello-world',
      clusterName: 'Default',
      namespace: 'default',
      provider: 'traefik',
      status: {
        phase: 'Succeeded',
        failedChecks: 1,
        canaryWeight: 0,
        iterations: 0,
        lastTransitionTime: '2022-05-11T13:54:51Z',
        conditions: [
          {
            type: 'Promoted',
            status: 'True',
            lastUpdateTime: '2022-05-11T13:54:51Z',
            lastTransitionTime: '2022-05-11T13:54:51Z',
            reason: 'Succeeded',
            message:
              'Canary analysis completed successfully, promotion finished.',
          },
        ],
      },
      targetDeployment: {
        uid: '4b871207-63e7-4981-b067-395c59b345254',
        resource_version: '',
      },
      targetReference: {
        kind: 'Deployment',
        name: 'hello-world',
      },
    },
    {
      name: 'podinfo',
      namespace: 'podinfo',
      clusterName: 'Default',
      provider: 'traefik',
      status: {
        phase: 'Progressing',
        failedChecks: 1,
        canaryWeight: 15,
        iterations: 0,
        lastTransitionTime: '2022-05-11T13:54:51Z',
        conditions: [
          {
            type: 'Promoted',
            status: 'True',
            lastUpdateTime: '2022-05-11T13:54:51Z',
            lastTransitionTime: '2022-05-11T13:54:51Z',
            reason: 'Progressing',
            message:
              'Canary analysis completed successfully, promotion finished.',
          },
        ],
      },
      targetDeployment: {
        uid: '4b871207-63e7-4981-b067-395csededd',
        resource_version: '',
      },
      targetReference: {
        kind: 'Deployment',
        name: 'hello-world',
      },
    },
  ];

  static getFlaggerStatus = (): Promise<any> => {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        resolve({
          Default: true,
          'LeafCluster-1': false,
          'LeafCluster-2': false,
          'LeafCluster-3': false,
        });
      }, 1000);
    });
  };

  static listCanaries = (): Promise<any> => {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        resolve({
          canaries: this.canariesArr,
          total: 4,
          nextPageToken: 'looooong token',
          errors: [],
        });
      }, 1000);
    });
  };

  static GetCanary = (id: string): Promise<any> => {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        resolve({
          canary:
            this.canariesArr[
              this.canariesArr.findIndex(e => e.targetDeployment.uid === id)
            ],
        });
      }, 1000);
    });
  };
}
