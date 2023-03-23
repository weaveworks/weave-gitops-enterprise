import { SecretDataType, SOPS } from ".";

export const convertToObject = (arr: any[]) => {
    const obj: any = {};
    arr.forEach(o => {
      obj[o.key] = o.value;
    });
    return obj;
  };
  export function scrollToAlertSection() {
    const element = document.getElementsByClassName('MuiAlert-root')[0];
    element?.scrollIntoView({ behavior: 'smooth' });
  }
  
  export const handelError = (err: any, setNotifications: any) => {
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
    const {
      clusterName,
      secretName,
      secretNamespace,
      kustomization,
      secretData,
      secretValue,
      secretType,
    } = formData;
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
    const data =
      secretType === SecretDataType.value
        ? {
            stringData: {
              string: secretValue,
            },
          }
        : { data: convertToObject(secretData) };
  
    return {
      encryptionPayload: {
        clusterName,
        name: secretName,
        namespace: secretNamespace,
        kustomization_name: k_name,
        kustomization_namespace: k_namespace,
        ...data,
      },
      cluster,
    };
  };