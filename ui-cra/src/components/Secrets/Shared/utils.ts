import { GitRepository } from '@weaveworks/weave-gitops';

export interface ExternalSecret {
  secretStoreType: any;
  secretPath: string | undefined;
  secretStore: string;
  dataSecretKey: string | undefined;
  secretStoreKind: string | undefined;
  secretStoreRef: string;
  clusterName: string;
  secretName: string;
  secretNamespace: string;
  includeAllProps: boolean;
  data: { id: number; key: string; value: string }[];

  repo: string | null | GitRepository;
  provider: string;
  branchName: string;
  pullRequestTitle: string;
  commitMessage: string;
  pullRequestDescription: string;
}
export function getESInitialData(
  callbackState: { state: { formData: ExternalSecret } } | null,
  random: string,
) {
  let defaultFormData = {
    repo: null,
    provider: '',
    branchName: `add-external-secret-branch-${random}`,
    pullRequestTitle: 'Add External Secret',
    commitMessage: 'Add External Secret',
    pullRequestDescription: 'This PR adds a new External Secret',
    dataSecretKey: '',
    secretStoreKind: '',
    secretStoreRef: '',
    clusterName: '',
    secretName: '',
    secretNamespace: '',
    secretStoreType: '',
    secretStore: '',
    secretPath: '',
    includeAllProps: false,
    data: [{ id: 1, key: '', value: '' }],
  };

  const initialFormData = {
    ...defaultFormData,
    ...callbackState?.state?.formData,
  };

  return { initialFormData };
}

export interface SOPS {
  clusterName: string;
  secretName: string;
  secretNamespace: string;
  encryptionType: string;
  kustomization: string;
  data: { id: number; key: string; value: string }[];
  repo: string | null | GitRepository;
  provider: string;
  branchName: string;
  pullRequestTitle: string;
  commitMessage: string;
  pullRequestDescription: string;
}
export function getInitialData(
  callbackState: { state: { formData: SOPS } } | null,
  random: string,
) {
  let defaultFormData = {
    repo: null,
    provider: '',
    branchName: `add-SOPS-secret-branch-${random}`,
    pullRequestTitle: 'Add SOPS Secret',
    commitMessage: 'Add SOPS Secret',
    pullRequestDescription: 'This PR adds a new SOPS Secret',
    clusterName: '',
    secretName: '',
    secretNamespace: '',
    encryptionType: 'GPG/AGE',
    kustomization: '',
    data: [{ id: 1, key: '', value: '' }],
  };

  const initialFormData = {
    ...defaultFormData,
    ...callbackState?.state?.formData,
  };

  return { initialFormData };
}
export const convertToObject = (
  arr: {
    key: string;
    value: string;
  }[],
) => {
  const obj: { [key: string]: string } = {};
  arr.forEach(o => {
    obj[o.key] = o.value;
  });
  return obj;
};
export function scrollToAlertSection() {
  const element = document.getElementsByClassName('MuiAlert-root')[0];
  element?.scrollIntoView({ behavior: 'smooth' });
}

export const handleError = (err: any, setNotifications: any) => {
  if (err.code === 401) {
    const { pathname, search } = window.location;
    const redirectUrl = encodeURIComponent(`${pathname}${search}`);
    const url = redirectUrl
      ? `/sign_in?redirect=${redirectUrl}`
      : `/sign_in?redirect=/`;
    window.location.href = url;
  }
  setNotifications([
    {
      message: { text: err.message },
      severity: 'error',
      display: 'top',
    },
  ]);
  scrollToAlertSection();
};

export const getFormattedPayload = (formData: SOPS) => {
  const { clusterName, secretName, secretNamespace, kustomization, data } =
    formData;
  const [k_name, k_namespace] = kustomization.split('/');
  const [c_namespace, c_name] = clusterName.split('/');
  const cluster =
    clusterName.split('/').length > 1
      ? {
          name: c_name,
          namespace: c_namespace,
        }
      : {
          name: c_namespace,
        };

  return {
    encryptionPayload: {
      clusterName,
      name: secretName,
      namespace: secretNamespace,
      kustomization_name: k_name,
      kustomization_namespace: k_namespace,
      data: convertToObject(data),
    },
    cluster,
  };
};
export const getESFormattedPayload = (formData: ExternalSecret) => {
  const {
    clusterName,
    secretName,
    secretNamespace,
    secretStoreRef,
    secretStoreKind,
    dataSecretKey,
    secretPath,
    includeAllProps,
    data,
  } = formData;
  const [c_namespace, c_name] = clusterName.split('/');
  const cluster =
    clusterName.split('/').length > 1
      ? {
          name: c_name,
          namespace: c_namespace,
        }
      : {
          name: c_namespace,
        };

  const dataObj = includeAllProps
    ? {
        dataFrom: {
          extract: {
            key: secretPath,
          },
        },
      }
    : {
        data: data.map(d => {
          return {
            secretKey: d.value,
            remoteRef: {
              key: d.key,
              property: d.value,
            },
          };
        }),
      };

  return {
    externalSecret: {
      metadata: {
        name: secretName,
        namespace: secretNamespace,
      },
      spec: {
        refreshInterval: '1h',
        secretStoreRef: {
          name: secretStoreRef,
          kind: secretStoreKind,
        },
        target: {
          name: dataSecretKey,
        },
        ...dataObj,
      },
    },
    cluster,
    isControlPlane: c_namespace ? true : false,
  };
};
