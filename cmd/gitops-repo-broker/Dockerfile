FROM docker.io/governmentpaas/git-ssh

# Add git provider hostnames to known hosts file so we can use
# StrickHostKeyChecking with git+ssh
ADD ./known_hosts.sh /home/wkp/known_hosts.sh
RUN bash /home/wkp/known_hosts.sh /etc/ssh/ssh_known_hosts && \
    rm /home/wkp/known_hosts.sh

ADD gitops-repo-broker /
ENTRYPOINT ["/gitops-repo-broker"]
