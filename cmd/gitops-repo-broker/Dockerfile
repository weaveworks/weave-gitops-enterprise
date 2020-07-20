FROM docker.io/governmentpaas/git-ssh

# Add git provider hostnames to known hosts file so we can use
# StrickHostKeyChecking with git+ssh
ADD ./known_hosts.sh /home/wkp/known_hosts.sh
RUN bash /home/wkp/known_hosts.sh /etc/ssh/ssh_known_hosts && \
    rm /home/wkp/known_hosts.sh

RUN wget https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.11.0/kubeseal-linux-amd64 -O kubeseal && \
    echo 'ae580e28400b4aa806adf3fd6b002799763820927bd6ac5a5df5250d50377e67  kubeseal' | sha256sum -c && \
    install -m 755 kubeseal /usr/local/bin/kubeseal

ADD gitops-repo-broker /
ENTRYPOINT ["/gitops-repo-broker"]
