# Third-party SDN integration

The default configuration for Neutron API service deployed by this operator is
to use ML2/OVN plugin. If you'd like to integrate Neutron with a different SDN,
the following steps may have to be followed.

TODO: check - does neutron depend on ovn-operator even when 3rd party SDN is
used? UPD: It does, so we should find a way to avoid ovn-operator deployment
and ovndbcluster creation for 3rd party setups. Report a jira for this.

## Controller nodes

### Build custom container image

To inject a 3rd party ML2 driver into neutron-server deployed by this operator, you should:

1. Build a custom image shipping neutron-server with your driver package
   installed and registered as a stevedore entrypoint.

   You can build your image on top of the default tcib build of neutron-server.
   https://github.com/openstack-k8s-operators/tcib/tree/main/container-images/tcib/base/os/neutron-base/neutron-server

2. Specify the new image to use for neutron service deployment.

   For this, set `containerImage` in Neutron spec.

### Enable custom ML2 drivers

To enable your ML2 driver for neutron-server, add this to your spec:

```
customServiceConfig: |
    [ml2]
    mechanism_drivers = <driver_name>
```

Note: you can add more options to the snippet, as needed. For example, since
the backend is not OVN, you may want to adjust `service_plugins` to exclude
`ovn-router` (which is part of the default list configured by the operator.)

```
customServiceConfig: |
    [DEFAULT]
    service_plugins=qos,router,trunk,segments,port_forwarding,log
    [ml2]
    mechanism_drivers = <driver_name>
```

### Communication to SDN backend

Depending on SDN of choice and its deployment, you may have to use one of the
following options to set Neutron service up for communication with the SDN
backend.

#### Internet communication

If your SDN interface is exposed via a Internet address, it is important that
the communication channel between the ML2 driver and the backend is secured. A
common approach to do so is to deploy SSL certificates to secure the channel.

Your ML2 driver may require presence of a certificate file on disk to use for
backend communication. In this case, consider preparing a
[Secret](https://kubernetes.io/docs/concepts/configuration/secret/) with the
contents of the certificate and mounting it as a file into the neutron-server
container.

Assuming Secret `backend-secret` contains the necessary certificate payload,
you can mount it into neutron-server container as follows:

```
spec:
  extraMounts:
    - extraVol:
      - volumes:
        - name: custom-secret
          volumeSource:
            secret:
              secretName: backend-secret
        mounts:
          - name: custom-mount
            mountPath: "/var/lib/neutron/third_party/backend.crt"
            readOnly: true
```

Note: Secret object preparation is out of scope for this document. You may use
https://cert-manager.io/docs/ or other mechanisms to manage your backend
certificates.

#### File system based communication

Your ml2 driver may have to talk to backend via a Unix socket file. Assuming
your backend is running in a pod too, perhaps through DaemonSet, this probably
means that your backend pod mounts a HostPath and then creates the socket file
there.

Regardless of that, your neutronapi container that runs your ml2 driver has to
get access to this socket file to talk to the backend. This can be achieved by
extraMounts with HostPath backend. Note that this will make your SCC
anyuid-hostmount. (TODO - check this is how it works.)

```
spec:
  extraMounts:
    - extraVol:
      - volumes:
        - name: custom-secret
          volumeSource:
            hostPath:
              path: "/var/run/backend/backend.sock"
        mounts:
          - name: custom-mount
            mountPath: "/var/run/backend/backend.sock"
```

NOTE: the same mechanism can be used to get access to other files hosted on the
node, if needed.

WARNING: don't use this mechanism to inject python code into the container,
instead see the section about building a custom image. Consult "Build custom
container image" section for the recommended mechanism to inject backend
specific python packages.

(Reasons - no compatibility guarantees between default image and your code,
untangled lifecycles between the default image and your code, no proper means
to deliver new code and detect its presence to reload the service if needed.)

## Compute nodes

While this operator may be used independently of dataplane-operator EDP nodes,
its regular environment implies EDP nodes. This means that you may need to
extend your backend configuration there.

### Define custom EDP Service

Your SDN may require deployment of backend specific services to nodes that are
not part of the Kubernetes cluster that runs the operator. Depending on how
these nodes are managed, you may apply corresponding changes elsewhere.

This operator is usually used in combination with
[dataplane-operator](https://github.com/openstack-k8s-operators/dataplane-operator).
You can refer to the
[following](https://openstack-k8s-operators.github.io/dataplane-operator/composable_services/#customizing-the-ansible-runner-image-used-by-a-service)
documentation describing the procedure to follow.
