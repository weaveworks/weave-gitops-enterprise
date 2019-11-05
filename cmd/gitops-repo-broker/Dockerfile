FROM fedora:30

RUN yum install -y git && yum clean all

# Add git hosts to known hosts file so we can use
# StrickHostKeyChecking with git+ssh
ADD ./known_hosts.sh /home/wkp/known_hosts.sh
RUN bash /home/wkp/known_hosts.sh /etc/ssh/ssh_known_hosts && \
    rm /home/wkp/known_hosts.sh


ADD gitops-repo-broker /
ENTRYPOINT ["/gitops-repo-broker"]
