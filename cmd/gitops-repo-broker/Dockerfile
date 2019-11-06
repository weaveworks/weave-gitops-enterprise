FROM docker.io/governmentpaas/git-ssh
COPY ./ssh_config /etc/ssh/ssh_config
ADD gitops-repo-broker /
ENTRYPOINT ["/gitops-repo-broker"]
