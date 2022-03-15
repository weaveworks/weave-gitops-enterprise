apiVersion: v1
kind: Config

clusters:
- name: ${cluster_name}
  cluster:
    server: https://${endpoint}
    certificate-authority-data: ${cluster_ca_certificate}

users:
- name: ${user_name}
  user:
    token: ${token}

contexts:
- context:
    cluster: ${cluster_name}
    user: ${user_name}
  name: ${context}
current-context: ${context}

