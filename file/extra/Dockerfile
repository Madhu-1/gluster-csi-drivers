FROM centos
MAINTAINER  Madhu Rajanna mrajanna@redhat.com
# Copy gluster-csi-driver from build  directory
COPY build/gluster-csi-driver /gluster-csi-driver

RUN yum -y install epel-release glusterfs-fuse && yum clean all

RUN rm -rf /var/cache/yum

ENTRYPOINT ["/gluster-csi-driver"]
