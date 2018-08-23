[![Build Status](https://travis-ci.org/Madhu-1/gluster-csi-drivers.svg?branch=master)](https://travis-ci.org/Madhu-1/gluster-csi-drivers)

## steps to setup kubernetes cluster and create pvc using glusterd2 and CSI drivers

## Installation of kubernets

* have atleast 4 vms to install kubernetes on it


  * 1 kubernetes master
  * 3 kubernetes nodes

#### Install Docker

In order to configure kubernetes cluster, it is require to install Docker. Execure below command to install Docker on all vm's.

```
#yum install -y docker
```

#### Enable & start Docker service.

```
#systemctl enable docker 
#systemctl start docker
```

#### Verify docker version is 1.12 and greater.

```
#docker version
```

### Disable SELinux:

```
#setenforce 0
```

In order for Kubernetes cluster to communicate internally, we have to disable SELinux.

#### Disable swap on all vms

```
# swapoff -a
```

#### Installing packages for Kubernetes:

On all vm's:

```
bash -c 'cat <<EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF'
```

```
#yum install -y kubelet kubeadm kubectl
```

#### Enable & start kubelet service:

```
#systemctl enable kubelet 
#systemctl start kubelet
```


Below settings will make sure IPTable is configured correctly.

Run this on all vm's
```
#bash -c 'cat <<EOF >  /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF'
```

##### Execute below command to apply above changes.

```
#sysctl --system
```

#### Disable and stop firewall on both master and nodes

```
#systemctl disable firewalld
```

```
systemctl stop firewalld
```

#### Configuring Kubernetes Master node
On your  master node execute below command:

```
#kubeadm init --pod-network-cidr 10.244.0.0/16
```

Once above command is completed, it will output Kubernetes cluster information. It will be needed to join worker nodes to Kubernetes cluster.

* Output from kubeadm init:

```
Your Kubernetes master has initialized successfully!

To start using your cluster, you need to run the following as a regular user:

  mkdir -p $HOME/.kube
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  sudo chown $(id -u):$(id -g) $HOME/.kube/config

You should now deploy a pod network to the cluster.
Run "kubectl apply -f [podnetwork].yaml" with one of the options listed at:
  https://kubernetes.io/docs/concepts/cluster-administration/addons/

You can now join any number of machines by running the following on each node
as root:

  kubeadm join 192.168.121.162:6443 --token paob5a.6h0ebbhapfycxz4k --discovery-token-ca-cert-hash sha256:6dab8d4f8b4dd3a301a1a15d5a4ae8d1bbacb541905732160a0c68db208a5ea9
```

#### On Master node apply below changes after kubeadm init successful configuration:

```
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

Configuring Pod Networking
Before we setup worker nodes, we need to ensure pod networking is functional. Pod networking is also a dependency for kube-dns pod to manage pod dns.

* Run this in kubernetes master node
```
#kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml
```
```
#kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/k8s-manifests/kube-flannel-rbac.yml
```
Ensure all pods are in running status by executing below command:

```
kubectl get pods --all-namespaces
```


Once all the pods are in running status, let's configure worker nodes.

Configure Kubernetes Worker nodes
To configure worker nodes to be part of Kubernetes cluster. We need to use kubeadm join command with token received from master node.

Execute below command to join worker node to Kubernetes Cluster.

```
#kubeadm join 192.168.121.162:6443 --token paob5a.6h0ebbhapfycxz4k --discovery-token-ca-cert-hash sha256:6dab8d4f8b4dd3a301a1a15d5a4ae8d1bbacb541905732160a0c68db208a5ea9
```

reponse from the above command will look like this

```
kubelet] Downloading configuration for the kubelet from the "kubelet-config-1.11" ConfigMap in the kube-system namespace
[kubelet] Writing kubelet configuration to file "/var/lib/kubelet/config.yaml"
[kubelet] Writing kubelet environment file with flags to file "/var/lib/kubelet/kubeadm-flags.env"
[preflight] Activating the kubelet service
[tlsbootstrap] Waiting for the kubelet to perform the TLS Bootstrap...
[patchnode] Uploading the CRI Socket information "/var/run/dockershim.sock" to the Node API object "node1" as an annotation

This node has joined the cluster:
* Certificate signing request was sent to master and a response
  was received.
* The Kubelet was informed of the new secure connection details.

Run 'kubectl get nodes' on the master to see this node join the cluster.

```
after running above command from all nodes 
check all the nodes have joined kubernetes cluster
```
kubectl get nodes
```

Output for above command

```

[root@master kubernetes]# kubectl get nodes
NAME      STATUS    ROLES     AGE       VERSION
master    Ready     master    1d        v1.11.1
node0     Ready     <none>    1d        v1.11.1
node1     Ready     <none>    1d        v1.11.1
node2     Ready     <none>    1d        v1.11.1
```


### deploying glusterd2 container in kubernetes

* copy deploy folder from this repo to the master node

Execute below command in master
```
kubectl create -f glusterd2.yml
```

once the glusterd2 container are up and running
we need to create a glusterd2 cluster

get all the node ips where gluster containers are running

```
[root@master kubernetes]# kubectl get po -o wide
NAME                                   READY     STATUS    RESTARTS   AGE       IP                NODE
glusterfs-42mtz                        1/1       Running   1          1d        192.168.121.198   node1
glusterfs-9cb6l                        1/1       Running   1          1d        192.168.121.186   node0
glusterfs-dj946                        1/1       Running   1          1d        192.168.121.114   node2

```

step inside any one of the gluster container 

```
kubectl exec -it glusterfs-42mtz /bin/bash
```

now do peer probing from one container to others

```
glustercli peer add <node-ip>

sample output:
[root@node1 /]# glustercli peer add 192.168.121.170
Peer add successful
+--------------------------------------+-------+-----------------------+-----------------------+
|                  ID                  | NAME  |   CLIENT ADDRESSES    |    PEER ADDRESSES     |
+--------------------------------------+-------+-----------------------+-----------------------+
| 708da5ec-285d-437e-8309-18502681e30d | node0 | 127.0.0.1:24007       | 192.168.121.170:24008 |
|                                      |       | 192.168.121.170:24007 |                       |
|                                      |       | 192.168.10.100:24007  |                       |
|                                      |       | 172.17.0.1:24007      |                       |
|                                      |       | 10.244.1.0:24007      |                       |
+--------------------------------------+-------+-----------------------+-----------------------+

```

get the peer id by below command
```
glustercli peer list
```

add device to gluster pods

```
glustercli device add <node-id> <device-name>

sample output:
[root@node1 /]# glustercli device add 1b82fb1b-471d-443f-bf4a-15c22a90e28b /dev/vdb
Device add successful
```

Execute below commands to deploy gluster CSI-driver on master

* deploying csi-attacher container

get rest auth secret from one of the gluster container to which glusterd2 container we will be sending the request to create volumes

rest auth file will be present in 
```
/usr/loca/var/lib/glusterd2/auth
```
update --secret and --glusterurl in 
csi-attacher-glusterfsplugin.yaml,csi-nodeplugin-glusterfsplugin.yaml and
csi-provisioner-glusterfsplugin.yaml

* --secret contains the content of auth file

* --glusterurl with Node IP where gluster pod is running which we are targeting to send request

```
kubectl create -f csi-attacher-glusterfsplugin.yaml
```
* creating rbac for csi-attacher 

```
kubectl create -f csi-attacher-rbac.yaml
```
* deploy csi node plugin 
```
kubectl create -f csi-nodeplugin-glusterfsplugin.yaml
```
* create rbac for nodeplugin

```
kubectl create -f csi-nodeplugin-rbac.yaml
```

* deploy provisioner 
```
kubectl create -f csi-provisioner-glusterfsplugin.yaml
```
* create rbac for csi provision
```
kubectl create -f csi-provisioner-rbac.yaml
```

now we have complete kubernetes setup with glusterd2 containers and csi-drivers

check all the container are up and running

```
[root@master vagrant]# kubectl get po
NAME                                   READY     STATUS    RESTARTS   AGE
csi-attacher-glusterfsplugin-0         2/2       Running   0          5m
csi-nodeplugin-glusterfsplugin-7btrv   2/2       Running   0          4m
csi-nodeplugin-glusterfsplugin-khcth   2/2       Running   0          4m
csi-nodeplugin-glusterfsplugin-thhss   2/2       Running   1          4m
csi-provisioner-glusterfsplugin-0      2/2       Running   0          4m
glusterfs-fdltf                        1/1       Running   1          29m
glusterfs-j6frk                        1/1       Running   1          29m
glusterfs-lq66c                        1/1       Running   1          29m

```


* create storage class

```
kubectl create -f storage.yml
```

* check storage class

```
[root@master vagrant]# kubectl get sc
NAME                     PROVISIONER             AGE
glusterfscsi (default)   org.gluster.glusterfs   2m

```
* create pvc 
```
kubectl create -f pvc.yml
```

* check pvc status 

```
[root@master ~]# kubectl get pvc
NAME          STATUS    VOLUME                 CAPACITY   ACCESS MODES   STORAGECLASS   AGE
glusterd2pv   Bound     pvc-1dc175c395a111e8   20Gi       RWX            glusterfscsi   3m

```

* create an app which uses this volume

```
kubectl create -f app.yml
```

* check created pvc is attached to pod

```
[root@master ~]# kubectl describe po/redis
Name:               redis
Namespace:          default
Priority:           0
PriorityClassName:  <none>
Node:               node0/192.168.121.170
Start Time:         Wed, 01 Aug 2018 15:43:46 +0000
Labels:             name=redis
Annotations:        <none>
Status:             Running
IP:                 10.244.1.3
Containers:
  redis:
    Container ID:   docker://184a1c2866908753439ff240492f4956f808182cba029c746866ca89cbb0910b
    Image:          redis
    Image ID:       docker-pullable://docker.io/redis@sha256:096cff9e6024603decb2915ea3e501c63c5bb241e1b56830a52acfd488873843
    Port:           <none>
    Host Port:      <none>
    State:          Running
      Started:      Wed, 01 Aug 2018 15:44:26 +0000
    Ready:          True
    Restart Count:  0
    Environment:    <none>
    Mounts:
      /mnt/gluster from glustercsivol (rw)
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-zlwqj (ro)
Conditions:
  Type              Status
  Initialized       True 
  Ready             True 
  ContainersReady   True 
  PodScheduled      True 
Volumes:
  glustercsivol:
    Type:       PersistentVolumeClaim (a reference to a PersistentVolumeClaim in the same namespace)
    ClaimName:  glusterd2pv
    ReadOnly:   false
  default-token-zlwqj:
    Type:        Secret (a volume populated by a Secret)
    SecretName:  default-token-zlwqj
    Optional:    false
QoS Class:       BestEffort
Node-Selectors:  <none>
Tolerations:     node.kubernetes.io/not-ready:NoExecute for 300s
                 node.kubernetes.io/unreachable:NoExecute for 300s
Events:
  Type    Reason                  Age   From                     Message
  ----    ------                  ----  ----                     -------
  Normal  Scheduled               1m    default-scheduler        Successfully assigned default/redis to node0
  Normal  SuccessfulAttachVolume  1m    attachdetach-controller  AttachVolume.Attach succeeded for volume "pvc-1dc175c395a111e8"
  Normal  Pulling                 59s   kubelet, node0           pulling image "redis"
  Normal  Pulled                  37s   kubelet, node0           Successfully pulled image "redis"
  Normal  Created                 37s   kubelet, node0           Created container
  Normal  Started                 37s   kubelet, node0           Started container

```


* TO delete app

```
[root@master working]# kubectl delete po/redis
pod "redis" deleted
```

* To delete pvc

```
[root@master working]# kubectl delete pvc/glusterd2pv
persistentvolumeclaim "glusterd2pv" deleted
```

* verification

do get pvc and pv check actually volume has deleted or not

```
[root@master working]# kubectl get pvc
No resources found.
```

```
[root@master working]# kubectl get pv
No resources found.
```
