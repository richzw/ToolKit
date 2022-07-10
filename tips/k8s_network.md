
- [K8S network 101](https://sookocheff.com/post/kubernetes/understanding-kubernetes-networking-model/)
  - K8S Basic
    ![img.png](k8s_network_arch.png)
    - API Server
      - everything is an API call served by the Kubernetes API server (kube-apiserver). 
      - The API server is a gateway to an etcd datastore that maintains the desired state of your application cluster.
    - Controllers
      - Once you’ve declared the desired state of your cluster using the API server, controllers ensure that the cluster’s current state matches the desired state by continuously watching the state of the API server and reacting to any changes.
      - Ex. when you create a new Pod using the API server, the Kubernetes scheduler (a controller) notices the change and makes a decision about where to place the Pod in the cluster. It then writes that state change using the API server (backed by etcd). 
      - The kubelet (a controller) then notices that new change and sets up the required networking functionality to make the Pod reachable within the cluster.
    - Scheduler
      - 调度程序是一个控制平面进程，它将 pod 分配给节点。它监视没有分配节点的新创建的 pod，并且对于调度程序发现的每个 pod，调度程序负责为该 pod 找到运行的最佳节点。
      - 调度程序不会指示所选节点运行 pod。Scheduler 所做的只是通过 API Server 更新 pod 定义。API server 通过 watch 机制通知 Kubelet pod 已经被调度。然后目标节点上的 kubelet 服务看到 pod 已被调度到它的节点，它创建并运行 pod 的容器。
    - Pod
      - A Pod is the atom of Kubernetes — the smallest deployable object for building applications.
      - A single Pod represents a running workload in your cluster and encapsulates one or more Docker containers, any required storage, and a unique IP address.
    - [Container](https://www.docker.com/resources/what-container/)
      - A container is a standard unit of software that packages up code and all its dependencies so the application runs quickly and reliably from one computing environment to another.
      - Cgroups - Limits and accounting of CPU memory network - configured by container runtime
      - Namespace - Isolation of process, CPU mount user network .. - configured by container runtime
      - ![img.png](k8s_network_container_cg.png)
    - Node
      - Nodes are the machines running the Kubernetes cluster. These can be bare metal, virtual machines, or anything else.
    - CNI
      - Some CNI plugins do a lot more than just ensuring Pods have IP addresses and that they can talk to each other
      ![img.png](k8s_network_summary.png)
      - [AWS CNI](https://github.com/aws/amazon-vpc-cni-k8s/blob/master/docs/cni-proposal.md)
      ![img.png](k8s_network_aws_pod2pod.png)
    - Services
      - because of the ephemeral nature of Pods, it is almost never a good idea to directly use Pod IP addresses. Pod IP addresses are not persisted across restarts and can change without warning, in response to events that cause Pod restarts (such as application failures, rollouts, rollbacks, scale-up/down, etc.).
      - Kubernetes Service objects allow you to assign a single virtual IP address to a set of Pods. Used to build service discovery. It works by keeping track of the state and IP addresses for a group of Pods and proxying / load-balancing traffic to them
      - Pods can also use internally available DNS names instead of Service IP addresses. This DNS system is powered by CoreDNS
      - Service的type类型
        - ClusterIP： 默认方式。根据是否生成ClusterIP又可分为普通Service和Headless Service两类：
          - 普通Service：通过为Kubernetes的Service分配一个集群内部可访问的固定虚拟IP（Cluster IP），实现集群内的访问。为最常见的方式。
          - Headless Service：该服务不会分配Cluster IP，也不通过kube-proxy做反向代理和负载均衡。而是通过DNS提供稳定的网络ID来访问，DNS会将headless service的后端直接解析为podIP列表。主要供StatefulSet中对应POD的序列用。
        - NodePort：除了使用Cluster IP之外，还通过将service的port映射到集群内每个节点的相同一个端口，实现通过nodeIP:nodePort从集群外访问服务。
        - LoadBalancer：和nodePort类似，不过除了使用一个Cluster IP和nodePort之外，还会向所使用的公有云申请一个负载均衡器，实现从集群外通过LB访问服务。在公有云提供的 Kubernetes 服务里，都使用了一个叫作 CloudProvider 的转接层，来跟公有云本身的 API 进行对接。所以，在上述 LoadBalancer 类型的 Service 被提交后，Kubernetes 就会调用 CloudProvider 在公有云上为你创建一个负载均衡服务，并且把被代理的 Pod 的 IP 地址配置给负载均衡服务做后端。
    - Endpoints
      - how do Services know which Pods to track, and which Pods are ready to accept traffic? The answer is Endpoints
      - ![img.png](k8s_network_endpoint.png)
      - Represent the list of IPs behind a Service
      - Recall that service had port and targetPort fields
    - DNS
      - Run as a pod in the cluster
      - Exposed by a Service VIP
      - Containers are configured by kubelet to use kube-dns
      - Default implementation is CoreDNS
    - Kubelet
      - Kubelet 是在集群中的每个节点上运行的代理，是负责在工作节点上运行的所有内容的组件。它确保容器在 Pod 中运行。
      - 通过在 API Server 中创建节点资源来注册它正在运行的节点。
      - 持续监控 API Server 上已调度到节点的 Pod。
      - 使用配置的容器运行时启动 pod 的容器。
      - 持续监控正在运行的容器并将其状态、事件和资源消耗报告给 API Server。
      - 运行容器活性探测，在探测失败时重新启动容器，在容器的 Pod 从 API Server 中删除时终止容器，并通知服务器 Pod 已终止。
    - Kube-proxy
      - Run on every Node in the cluster
      - 它负责监视 API Server 以了解Service和 pod 定义的更改，以保持整个网络配置的最新状态。当一个Service由多个 pod 时，proxy会在这些 pod 之间负载平衡。
      - kube-proxy 之所以得名，是因为它是一个实际的代理服务器，用于接受连接并将它们代理到 Pod，当前的实现使用 iptables 或 ipvs 规则将数据包重定向到随机选择的后端 Pod，而不通过实际的代理服务器传递它们。
      - Watch Service and Endpoints, link Endpoints(backends) with Service(frontends)
  - K8S network requirement
    - all pods can communicate with all other pods without using network address translation (NAT)
    - all Nodes can communicate with all Pods without NAT
    - the IP that a Pod sees itself as is the same IP that others see it as
  - Container to Container network
    ![img.png](k8s_network_pod.png)
    - Containers in a pod has the same network namespace
      - they have same network configuration
      - sharing the same Pod IP address
    - network accessing via loopback or eth0 interface
    - package are always handled in the network namespace
    ![img.png](k8s_network_container.png)
  - Pod to Pod network
    - every Pod has a real IP address and each Pod communicates with other Pods using that IP address.
    - namespaces can be connected using a Linux `Virtual Ethernet Device` or `veth pair` consisting of two virtual interfaces that can be spread over multiple namespaces.
    - A Linux Ethernet bridge is a virtual Layer 2 networking device used to unite two or more network segments, working transparently to connect two networks together.
    - Bridges implement the ARP protocol to discover the link-layer MAC address associated with a given IP address.
    ![img.png](k8s_network_pod2pod.png)
    ![img.png](k8s_network_pod2pod_acrossnode.png)

    |  | L2 | Route | Overlay | Cloud |
    | --- | --- | --- | --- | --- |
    | Summary | Pods Communicate using L2 | Pods traffic is routed in underlay network | Pod traffic is encapsulated and use underlay for reachability | Pod traffic is routed in cloud virtual network |
    | Underlying Tech | L2 ARP, broadcast | - Routing protocoal - BGP | VxLan, UDP encapluation in user space | Pre-programmed fabric using controller |
    | Ex. | Pod 2 Pod on the same node | - Calico - Flannel(HostGW) | - Flannel - Weave | - GKE - EKS |
    - Overlay network
      - 它是指构建在另一个网络上的计算机网络，这是一种网络虚拟化技术的形式. Overlay 底层依赖的网络就是 Underlay 网络，这两个概念也经常成对出现
      - Underlay 网络是专门用来承载用户 IP 流量的基础架构层，它与 Overlay 网络之间的关系有点类似物理机和虚拟机
      - 在实践中我们一般会使用虚拟局域网扩展技术（Virtual Extensible LAN，VxLAN）组建 Overlay 网络。在下图中，两个物理机可以通过三层的 IP 网络互相访问
      ![img_1.png](k8s_network_vxlan.png)
      - VxLAN 使用虚拟隧道端点（Virtual Tunnel End Point、VTEP）设备对服务器发出和收到的数据包进行二次封装和解封。
      - 虚拟网络标识符（VxLAN Network Identifier、VNI）, VxLAN 会使用 24 比特的 VNI 表示虚拟网络个数，总共可以表示 16,777,216 个虚拟网络，这也就能满足数据中心多租户网络隔离的需求了。
      ![img.png](k8s_network_vxlan_frame.png)
      ![img.png](k8s_network_overlay_packet.png)
    - Inside Pod
      ```shell
      IP address
      
      # ip addr show
      1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1
         link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
         inet 127.0.0.1/8 scope host lo
            valid_lft forever preferred_lft forever
         inet6 ::1/128 scope host 
            valid_lft forever preferred_lft forever
      3: eth0@if231: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue state UP 
         link/ether 56:41:95:26:17:41 brd ff:ff:ff:ff:ff:ff
         inet 10.0.97.30/32 brd 10.0.97.226 scope global eth0 <<<<<<< ENI's secondary IP address
            valid_lft forever preferred_lft forever
         inet6 fe80::5441:95ff:fe26:1741/64 scope link 
            valid_lft forever preferred_lft forever
      routes
      
      # ip route show
      default via 169.254.1.1 dev eth0 
      169.254.1.1 dev eth0 
      static arp
      
      # arp -a
      ? (169.254.1.1) at 2a:09:74:cd:c4:62 [ether] PERM on eth0
      ```
    - On the Node
      ```shell
      - There are multiple routing tables used to route incoming/outgoing Pod's traffic.
      
      - main (toPod) route table is used to route to Pod traffic
      # ip route show
      default via 10.0.96.1 dev eth0 
      10.0.96.0/19 dev eth0  proto kernel  scope link  src 10.0.104.183 
      10.0.97.30 dev aws8db0408c9a8  scope link  <------------------------Pod's IP
      10.0.97.159 dev awsbcd978401eb  scope link 
      10.0.97.226 dev awsc2f87dc4cdd  scope link 
      10.0.102.98 dev aws4914061689b  scope link 
      ...
      - Each ENI has its own route table which is used to route pod's outgoing traffic, where pod is allocated with one of the ENI's secondary IP address
      # ip route show table eni-1
      default via 10.0.96.1 dev eth1 
      10.0.96.1 dev eth1  scope link 
      - Here is the routing rules to enforce policy routing
      # ip rule list
      0:	from all lookup local 
      512:	from all to 10.0.97.30 lookup main <---------- to Pod's traffic
      1025:	not from all to 10.0.0.0/16 lookup main 
      1536:	from 10.0.97.30 lookup eni-1 <-------------- from Pod's traffic
      ```
  - Pod to Service network
    - Pod IP address - are mutable and will appear and disappear due to scaling up or down
    - Service assign a single VIP for load balance between a group of
    - [Kube-Proxy](https://mayankshah.dev/blog/demystifying-kube-proxy/)
      - a network proxy that runs on each node in your cluster. It watches Service and Endpoints objects and accordingly updates the routing rules on its host nodes to allow communicating over Services.
      - User Model
        - userland TCP/UDP proxy
      - IPtables
        - User IPtables to load-balance traffic
      - IPVS - IPVS (IP Virtual Server) implements transport-layer load balancing, usually called Layer 4 LAN switching, as part of Linux kernel.
        - User kernel LVS
        - Faster than IPtables
      - [IPVS vs IPTABLES](https://github.com/kubernetes/kubernetes/blob/master/pkg/proxy/ipvs/README.md)
        - Both IPVS and IPTABLES are based on netfilter. 
        - Differences between IPVS mode and IPTABLES mode are as follows:
          - IPVS provides better scalability and performance for large clusters.
          - IPVS supports more sophisticated load balancing algorithms than IPTABLES (least load, least connections, locality, weighted, etc.).
          - IPVS supports server health checking and connection retries, etc.
    - Using DNS
      - Kubernetes can optionally use DNS to avoid having to hard-code a Service’s cluster IP address into your application.
      - It configures the kubelets running on each Node so that containers use the DNS Service’s IP to resolve DNS names.
      - A DNS Pod consists of three separate containers:
        - kubedns: watches the Kubernetes master for changes in Services and Endpoints, and maintains in-memory lookup structures to serve DNS requests.
        - dnsmasq: adds DNS caching to improve performance.
        - sidecar: provides a single health check endpoint to perform healthchecks for dnsmasq and kubedns.
    - Service
        ```shell
        
        > kc get svc -n kube-system
        NAME                                TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                        AGE
        kube-dns                            ClusterIP   10.100.0.10      <none>        53/UDP,53/TCP                  25d
        
        > kc get ep -n beta
        NAME                               ENDPOINTS                                                                 AGE
        guild-beta-home-design             172.20.250.219:8080,172.20.251.121:8080                                   21d
        guild-beta-home-design-local       172.20.250.219:9420,172.20.251.121:9420                                   21d
        
        >  kc exec -it commgame-beta-word-connect-6cdbbc8486-qww9v -n beta -- sh
        / # nslookup 172.20.250.219 10.100.0.10
        Server:    10.100.0.10
        Address 1: 10.100.0.10 kube-dns.kube-system.svc.cluster.local
        
        Name:      172.20.250.219
        Address 1: 172.20.250.219 172-20-250-219.guild-beta-home-design.beta.svc.cluster.local
        / # nslookup idgenerator-local.beta.svc.cluster.local
        nslookup: can't resolve '(null)': Name does not resolve
        
        Name:      idgenerator-local.beta.svc.cluster.local
        Address 1: 172.20.250.157 172-20-250-157.idgenerator-local.beta.svc.cluster.local
        ```
  - Internet to Service network
    - Layer 4
      - NodePort
        - Node IP is used for external communication
        - Service is exposed using a reserved port in all nodes of cluster
      - Loadbalance
        - Each service needs to have own external IP
        - Typically implemented as NLB
    ![img.png](k8s_network_alb.png)
    - Layer 7
      ```shell
      > kc get svc -n beta -o wide
      NAME             TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)        AGE    SELECTOR
      word-v5-beta     NodePort    10.100.68.182    <none>        80:32723/TCP   283d   app=word-v5,env=be
      
      > kc get endpoints -n beta -o wide
      NAME             ENDPOINTS                                AGE
      word-v5-beta     172.20.254.146:3021                      283d
      
      > kc get po -n beta -o wide
      NAME                                  READY   STATUS    RESTARTS   AGE     IP               NODE                             NOMINATED NODE   READINESS GATES
      word-v5-beta-8555cc6dcd-7vbkl         1/1     Running   0          7d4h    172.20.254.146   ip-172-20-254-241.ec2.internal   <none>           <none>
      ```
      ```shell
      > iptables -t nat -nL
      
      Chain KUBE-NODEPORTS (1 references)
      target     prot opt source               destination
      KUBE-MARK-MASQ  tcp  --  0.0.0.0/0            0.0.0.0/0            /* beta/word-v5-beta:http */ tcp dpt:32723
      KUBE-SVC-7EUSQI7QIASQWEEM  tcp  --  0.0.0.0/0            0.0.0.0/0            /* beta/word-v5-beta:http */ tcp dpt:32723
      
      Chain KUBE-SERVICES (2 references)
      target     prot opt source               destination
      KUBE-SVC-7EUSQI7QIASQWEEM  tcp  --  0.0.0.0/0            10.100.68.182        /* beta/word-v5-beta:http cluster IP */ tcp dpt:80
      
      Chain KUBE-SVC-7EUSQI7QIASQWEEM (2 references)
      target     prot opt source               destination
      KUBE-SEP-W5KHKXQHUD4WIQCU  all  --  0.0.0.0/0            0.0.0.0/0            /* beta/word-v5-beta:http */
      
      Chain KUBE-SEP-W5KHKXQHUD4WIQCU (1 references)
      target     prot opt source               destination
      KUBE-MARK-MASQ  all  --  172.20.254.146       0.0.0.0/0            /* beta/word-v5-beta:http */
      DNAT       tcp  --  0.0.0.0/0            0.0.0.0/0            /* beta/word-v5-beta:http */ tcp to:172.20.254.146:3021
      ```


- [CNI](https://platform9.com/blog/the-ultimate-guide-to-using-calico-flannel-weave-and-cilium/)
  - flannel
    - Flannel runs a simple overlay network across all the nodes of the Kubernetes cluster. 
    - It provides networking at Layer 3, the Network Layer of the OSI networking model. 
    - Flannel supports [VXLAN](https://support.huawei.com/enterprise/zh/doc/EDOC1100087027) as its default backend, although you can also configure it to use UDP and host-gw. 
    - Some experimental backends like AWS VPC, AliVPC, IPIP, and IPSec are also available, but not officially supported at present.
    - One of the drawbacks of Flannel is its lack of advanced features, such as the ability to configure network policies and firewalls.
  - Calico
    - Calico operates on Layer 3 of the OSI model and uses the BGP protocol to move network packets between nodes in its default configuration with IP in IP for encapsulation. 
    - Using BGP, Calico directs packets natively, without needing to wrap them in additional layers of encapsulation. 
    - This approach improves performance and simplifies troubleshooting network problems compared with more complex backends, like VXLAN.
    - Calico’s most valuable feature is its support for network policies. By defining and enforcing network policies, you can prescribe which pods can send and receive traffic and manage security within the network.
  - Weave
    - Weave creates a mesh overlay between all nodes of a Kubernetes cluster and uses this in combination with a routing component on each node to dynamically route traffic throughout the cluster. By default, Weave routes packets using the fast datapath method, which attempts to send traffic between nodes along the shortest path.
    - Weave includes features such as creating and enforcing network policies and allows you to configure encryption for your entire network. If configured, Weave uses NaCl encryption for sleeve traffic and IPsec ESP encryption for fast datapath traffic.
  - Cilium
    - A relative newcomer to the land of Kubernetes CNI plugins is Cilium. Cilium and its observability tool, Hubble, take advantage of eBPF.


- [LVS](https://new.qq.com/omn/20200718/20200718A05H2H00.html)： 
  - LVS是Linux Virtual Server的简写，也就是Linux 虚拟服务器，是一个虚拟的服务器集群系统.
    通过 LVS 达到的负载均衡技术和 Linux 操作系统实现一个高性能高可用的 Linux 服务器集群，具有良好的可靠性、可扩展性和可操作性
  - LVS 与 Nginx 功能对比
    - LVS 比 Nginx 具有更强的抗负载能力，性能高，对内存和 CPU 资源消耗较低；
    - LVS 工作在网络层，具体流量由操作系统内核进行处理，Nginx 工作在应用层，可针对 HTTP 应用实施一些分流策略；
    - LVS 安装配置较复杂，网络依赖性大，稳定性高。Nginx 安装配置较简单，网络依赖性小；
    - LVS 不支持正则匹配处理，无法实现动静分离效果。
    - LVS 适用的协议范围广。Nginx 仅支持 HTTP、HTTPS、Email 协议，适用范围小；
  - LVS 由两部分程序组成，包括 ipvs 和 ipvsadm
    - ipvs(ip virtual server)：LVS 是基于内核态的 netfilter 框架实现的 IPVS 功能，工作在内核态。用户配置 VIP 等相关信息并传递到 IPVS 就需要用到 ipvsadm 工具。
      - iptables 是位于用户空间，而 Netfilter 是位于内核空间。iptables 只是用户空间编写和传递规则的工具而已，真正工作的还是 netfilter
      - LVS 基于 netfilter 框架，工作在 INPUT 链上，在 INPUT 链上注册 ip_vs_in HOOK 函数，进行 IPVS 相关主流程
    - ipvsadm：ipvsadm 是 LVS 用户态的配套工具，可以实现 VIP 和 RS 的增删改查功能，是基于 netlink 或 raw socket 方式与内核 LVS 进行通信的，如果 LVS 类比于 netfilter，那 ipvsadm 就是类似 iptables 工具的地位。
  - LVS 负载均衡的三种工作模式
    - 地址转换（NAT）
      - 类似于防火墙的私有网络结构，负载调度器作为所有服务器节点的网关，作为客户机的访问入口，也是各节点回应客户机的访问出口，服务器节点使用私有 IP 地址，与负载调度器位于同一个物理网络，安全性要优于其他两种方式。
      - 优点：
      - 支持 Windows 操作系统；
      - 支持端口映射，如 RS 服务器 PORT 与 VPORT 不一致的话，LVS 会修改目的 IP 地址和 DPORT 以支持端口映射；
      - 缺点：
      - RS 服务器需配置网关；
      - 双向流量对 LVS 会产生较大的负载压力；
    - IP 隧道（TUN）
      - 采用开放式的网络结构，负载调度器作为客户机的访问入口，各节点通过各自的 Internet 连接直接回应给客户机，而不经过负载调度器，服务器节点分散在互联网中的不同位置，有独立的公网 IP 地址，通过专用 IP 隧道与负载调度器相互通信。
      - 优点：
        - 单臂模式，LVS 负载压力小；
        - 数据包修改小，信息完整性高；
        - 可跨机房；
      - 缺点：
        - 不支持端口映射；
        - 需在 RS 后端服务器安装模块及配置 VIP；
        - 隧道头部 IP 地址固定，RS 后端服务器网卡可能会不均匀；
        - 隧道头部的加入可能会导致分片，最终会影响服务器性能；
    - 直接路由（DR）
      - 采用半开放式的网络结构，与 TUN 模式的结构类似，但各节点并不是分散在各个地方，而是与调度器位于同一个物理网络，负载调度器与各节点服务器通过本地网络连接，不需要建立专用的 IP 隧道。它是最常用的工作模式，因为它的功能性强大。
      - 优点：
        - 响应数据不经过 LVS，性能高；
        - 对数据包修改小，信息完整性好；
      - 缺点：
        - LVS 与 RS 必须在同一个物理网络；
        - RS 上必须配置 lo 和其他内核参数；
        - 不支持端口映射；
  - LVS 的十种负载调度算法
    - 轮询：Round Robin，将收到的访问请求按顺序轮流分配给群集中的各节点真实服务器中，不管服务器实际的连接数和系统负载。
    - 加权轮询：Weighted Round
    - Robin，根据真实服务器的处理能力轮流分配收到的访问请求，调度器可自动查询各节点的负载情况，并动态跳转其权重，保证处理能力强的服务器承担更多的访问量。
    - 最少连接：Least Connections，根据真实服务器已建立的连接数进行分配，将收到的访问请求优先分配给连接数少的节点，如所有服务器节点性能都均衡，可采用这种方式更好的均衡负载。
    - 加权最少连接：Weighted Least Connections，服务器节点的性能差异较大的情况下，可以为真实服务器自动调整权重，权重较高的节点将承担更大的活动连接负载。
    - 基于局部性的最少连接：LBLC，基于局部性的最少连接调度算法用于目标 IP 负载平衡，通常在高速缓存群集中使用。如服务器处于活动状态且处于负载状态，此算法通常会将发往 IP
    - 地址的数据包定向到其服务器。如果服务器超载（其活动连接数大于其权重），并且服务器处于半负载状态，则将加权最少连接服务器分配给该 IP 地址。
    - 复杂的基于局部性的最少连接：LBLCR，具有复杂调度算法的基于位置的最少连接也用于目标IP负载平衡，通常在高速缓存群集中使用。与 LBLC 调度有以下不同：负载平衡器维护从目标到可以
    - 目标提供服务的一组服务器节点的映射。对目标的请求将分配给目标服务器集中的最少连接节点。如果服务器集中的所有节点都超载，则它将拾取群集中的最少连接节点，并将其添加到目标服务
    - 群中。如果在指定时间内未修改服务器集群，则从服务器集群中删除负载最大的节点，以避免高度负载。
    - 目标地址散列调度算法：DH，该算法是根据目标 IP 地址通过散列函数将目标 IP 与服务器建立映射关系，出现服务器不可用或负载过高的情况下，发往该目标 IP 的请求会固定发给该服务器。
    - 源地址散列调度算法：SH，与目标地址散列调度算法类似，但它是根据源地址散列算法进行静态分配固定的服务器资源。
    - 最短延迟调度：SED，最短的预期延迟调度算法将网络连接分配给具有最短的预期延迟的服务器。如果将请求发送到第 i 个服务器，则预期的延迟时间为（Ci +1）/ Ui，其中 Ci 是第 i
    - 个服务器上的连接数，而 Ui 是第 i 个服务器的固定服务速率（权重） 。
    - 永不排队调度：NQ，从不队列调度算法采用两速模型。当有空闲服务器可用时，请求会发送到空闲服务器，而不是等待快速响应的服务器。如果没有可用的空闲服务器，则请求将被发送到服务器，以使其预期延迟最小化（最短预期延迟调度算法）

- [tunneling](https://wiki.linuxfoundation.org/networking/tunneling)
  - Tunneling is a way to transform data frames to allow them pass networks with incompatible address spaces or even incompatible protocols.
  - Linux kernel supports 3 tunnel types: 
    - IPIP (IPv4 in IPv4)
      - It has the lowest overhead, but can incapsulate only IPv4 unicast traffic, so you will not be able to setup OSPF, RIP or any other multicast-based protocol.
      - You can setup only one tunnel for unique tunnel endpoints pair
    - GRE (IPv4/IPv6 over IPv4) 
      - GRE tunnels can incapsulate IPv4/IPv6 unicast/multicast traffic, so it is de-facto tunnel standard for dynamic routed networks.
      - You can setup up to 64K tunnels for an unique tunnel endpoints pair
    - SIT (IPv6 over IPv4)
      - SIT stands for Simple Internet Transition. Its main purpose is to interconnect isolated IPv6 networks, located in global IPv4 Internet.
      - SIT works like IPIP. Once loaded, ipv6 module can't be unloaded.
      - You can get your own IPv6 prefix and a SIT tunnel from a tunnel broker.

- [A Deep Dive into Iptables and Netfilter Architecture](https://www.digitalocean.com/community/tutorials/a-deep-dive-into-iptables-and-netfilter-architecture)
  - What Are IPTables and Netfilter
    - The basic firewall software most commonly used in Linux is called iptables. The iptables firewall works by interacting with the packet filtering hooks in the Linux kernel’s networking stack. These kernel hooks are known as the netfilter framework.
  - Netfilter Hooks
    - There are five netfilter hooks that programs can register with.
    - NF_IP_PRE_ROUTING: This hook will be triggered by any incoming traffic very soon after entering the network stack. This hook is processed before any routing decisions have been made regarding where to send the packet.
    - NF_IP_LOCAL_IN: This hook is triggered after an incoming packet has been routed if the packet is destined for the local system.
    - NF_IP_FORWARD: This hook is triggered after an incoming packet has been routed if the packet is to be forwarded to another host.
    - NF_IP_LOCAL_OUT: This hook is triggered by any locally created outbound traffic as soon it hits the network stack.
    - NF_IP_POST_ROUTING: This hook is triggered by any outgoing or forwarded traffic after routing has taken place and just before being put out on the wire.
  - IPTable tables and chains
    ![img.png](k8s_network_iptables.png)
    - The iptables firewall uses tables to organize its rules
    - Within each iptables table, rules are further organized within separate “chains”. 
    - the built-in chains represent the netfilter hooks which trigger them. Chains basically determine when rules will be evaluated.
  - Which Tables are Available
    - The filter table is one of the most widely used tables in iptables. The filter table is used to make decisions about whether to let a packet continue to its intended destination or to deny its request.
    - The nat table is used to implement network address translation rules.
    - The mangle table is used to alter the IP headers of the packet in various ways.
    - The raw table has a very narrowly defined function. Its only purpose is to provide a mechanism for marking packets in order to opt-out of connection tracking.
  - Which Chains are Implemented in Each Table
    ![img.png](k8s_network_iptable_chain.png)

- [Node Proxy](https://cloudnative.to/blog/k8s-node-proxy/)
  - Netfilter
    - 主机上的所有数据包都将通过 netfilter 框架
    - 在 netfilter 框架中有 5 个钩子点：PRE_ROUTING, INPUT, FORWARD, OUTPUT, POST_ROUTING
    - 命令行工具 iptables 可用于动态地将规则插入到钩子点中
    - 可以通过组合各种 iptables 规则来操作数据包
      - filter：做正常的过滤，如接受，拒绝/删，跳
      - nat：网络地址转换，包括 SNAT（源 nat) 和 DNAT（目的 nat)
      - mangle：修改包属性，例如 TTL
      - raw：最早的处理点，连接跟踪前的特殊处理 (conntrack 或 CT，也包含在上图中，但这不是链）
      - security
  - Cross-host 网络模型
    - 主机 A 上的实例（容器、VM 等）如何与主机 B 上的另一个实例通信？有很多解决方案：
      - 直接路由：BGP 等
      - 隧道：VxLAN, IPIP, GRE 等
      - NAT：例如 docker 的桥接网络模式
  - Service
    - Service 是一种抽象，它定义了一组 pod 的逻辑集和访问它们的策略。
      - ClusterIP：通过 VIP 访问 Service，但该 VIP 只能在此集群内访问
        - 对 ClusterIP 的一个常见误解是，ClusterIP 是可访问的——它们不是通过定义访问的。如果 ping 一个 ClusterIP，可能会发现它不可访问。
        - 根据定义，<Protocol,ClusterIP,Port> 元组独特地定义了一个服务（因此也定义了一个拦截规则）。例如，如果一个服务被定义为 <tcp,10.7.0.100,80>，那么代理只处理 tcp:10.7.0.100:80 的流量，其他流量，例如。tcp:10.7.0.100:8080, udp:10.7.0.100:80 将不会被代理。因此，也无法访问 ClusterIP（ICMP 流量）。
        - 但是，如果你使用的是带有 IPVS 模式的 kube-proxy，那么确实可以通过 ping 访问 ClusterIP。这是因为 IPVS 模式实现比定义所需要的做得更多。
      - NodePort：通过 NodeIP:NodePort 访问 Service，这意味着该端口将暴露在集群内的所有节点上
      - ExternalIP：与 ClusterIP 相同，但是这个 VIP 可以从这个集群之外访问
      - LoadBalancer
    - 一个 Service 有一个 VIP（本文中的 ClusterIP）和多个端点（后端 pod）。每个 pod 或节点都可以通过 VIP 直接访问应用程序。要做到这一点，节点代理程序需要在每个节点上运行，它应该能够透明地拦截到任何 ClusterIP:Port的流量，并将它们重定向到一个或多个后端 pod。

- [Docker网络原理](https://mp.weixin.qq.com/s/jJiX47kRTfX-3UnbN8cvtQ)
  - Linux veth pair
    - veth pair 是成对出现的一种虚拟网络设备接口，一端连着网络协议栈，一端彼此相连
  - Docker0
    - lo和eth0在我们的虚拟机启动的时候就会创建，但是docker0在我们安装了docker的时候就会创建。docker0用来和虚拟机之间通信
    - 我们每启动一个容器，就会多出一对网卡，同时他们被连接到docker0上，而docker0又和虚拟机之间连通。
  - ![img.png](k8s_network_docker_network.png)

- [X.509 Encodings and Conversions](https://www.ssl.com/guide/pem-der-crt-and-cer-x-509-encodings-and-conversions/)
  - You may have seen digital certificate files with a variety of filename extensions, such as .crt, .cer, .pem, or .der. These extensions generally map to two major encoding schemes for X.509 certificates and keys: PEM (Base64 ASCII), and DER (binary). 
  - PEM
    - PEM (originally “Privacy Enhanced Mail”) is the most common format for X.509 certificates, CSRs, and cryptographic keys. A PEM file is a text file containing one or more items in Base64 ASCII encoding, each with plain-text headers and footers (e.g. -----BEGIN CERTIFICATE----- and -----END CERTIFICATE-----)
    - PEM Filename Extensions - .crt, .pem, .cer, .key (for private keys), and .ca-bundle
    - View contents of PEM certificate file `openssl x509 -in CERTIFICATE.pem -text -noout `
    - Convert PEM certificate to DER `openssl x509 -outform der -in CERTIFICATE.pem -out CERTIFICATE.der`
    - Convert PEM certificate with chain of trust to PKCS#7
      - PKCS#7 (also known as P7B) is a container format for digital certificates that is most often found in Windows and Java server contexts, and usually has the extension .p7b. PKCS#7 files are not used to store private keys. In the example below, -certfile MORE.pem represents a file with chained intermediate and root certificates (such as a .ca-bundle file downloaded from SSL.com).
      - `openssl crl2pkcs7 -nocrl -certfile CERTIFICATE.pem -certfile MORE.pem -out CERTIFICATE.p7b`
    - Convert PEM certificate with chain of trust and private key to PKCS#12
      - PKCS#12 (also known as PKCS12 or PFX) is a common binary format for storing a certificate chain and private key in a single, encryptable file, and usually have the filename extensions .p12 or .pfx. In the example below, -certfile MORE.pem adds a file with chained intermediate and root certificates (such as a .ca-bundle file downloaded from SSL.com), and -inkey PRIVATEKEY.key adds the private key for CERTIFICATE.crt(the end-entity certificate). Please see this how-to for a more detailed explanation of the command shown.
      - `openssl pkcs12 -export -out CERTIFICATE.pfx -inkey PRIVATEKEY.key -in CERTIFICATE.crt -certfile MORE.crt`
  - DER
    - DER (Distinguished Encoding Rules) is a binary encoding for X.509 certificates and private keys. Unlike PEM, DER-encoded files do not contain plain text statements such as -----BEGIN CERTIFICATE-----.
    - DER Filename Extensions - .der and .cer.
    - View contents of DER-encoded certificate file  `openssl x509 -inform der -in CERTIFICATE.der -text -noout`
    - Convert DER-encoded certificate to PEM `openssl x509 -inform der -in CERTIFICATE.der -out CERTIFICATE.pem`
- [TLS 单向和双向认证](https://mp.weixin.qq.com/s/JOpega3ud9P7NDNsGAwCCg)
  - SSL 证书 （也称为 TLS 或 SSL /TLS 证书）是将网站的身份绑定到由公共密钥和私有密钥组成的加密密钥对的数字文档。 证书中包含的公钥允许 Web 浏览器执行以下操作： 通过 TLS 和 HTTPS 协议。 私钥在服务器上保持安全，并用于对网页和其他文档（例如图像和 JavaScript 文件）进行数字签名。
  - 双向 TLS（mTLS)
    - TLS 服务器端提供一个授信证书，当我们使用 https 协议访问服务器端时，客户端会向服务器端索取证书并认证（浏览器会与自己的授信域匹配或弹出不安全的页面）。
    - mTLS 则是由同一个 Root CA 生成两套证书，即客户端证书和服务端证书。客户端使用 https 访问服务端时，双方会交换证书，并进行认证，认证通过方可通信。
  - 证书格式类型
    - .DER .CER，文件是二进制格式，只保存证书，不保存私钥。
    - .PEM，一般是文本格式(Base64 ASCII），可保存证书，可保存私钥。
    - .CRT，可以是二进制格式，可以是文本格式，与 .DER 格式相同，不保存私钥。
    - .PFX .P12，二进制格式，同时包含证书和私钥，一般有密码保护。
    - .JKS，二进制格式，同时包含证书和私钥，一般有密码保护。
  - 证书生成
    ```shell
    # 生成CA的私钥和证书
    echo Generate the ca certificate
    openssl genrsa -out ../certs/ca.key 4096
    openssl req -x509 -sha256 -new -nodes -key ../certs/ca.key -days 3650 -subj "/C=IN/ST=UK/L=Dehradun/O=VMware/CN=Hemant Root CA" -extensions v3_ca -out ../certs/ca.crt
    
    # 生成服务端的私钥和证书
    echo generating server certificate
    openssl genrsa -out ../certs/server.key 2048
    openssl req -new -subj "/C=IN/ST=UK/L=Dehradun/O=VMware/CN=localhost" -key ../certs/server.key -out server_signing_req.csr
    openssl x509 -req -days 365 -in server_signing_req.csr -CA ../certs/ca.crt -CAkey ../certs/ca.key -CAcreateserial -out ../certs/server.crt
    del server_signing_req.csr
    
    # 生成客户端的私钥和证书
    echo generating client certificate
    openssl genrsa -out ../certs/client.key 2048
    openssl req -new -subj "/C=IN/ST=UK/L=Dehradun/O=VMware/CN=localhost" -key ../certs/client.key -out client_signing_req.csr
    openssl x509 -req -days 365 -in client_signing_req.csr -CA ../certs/ca.crt -CAkey ../certs/ca.key -CAcreateserial -out ../certs/client.crt
    rm client_signing_req.csr
    
    # 验证证书
    openssl verify -CAfile ../certs/ca.crt ../certs/server.crt
    openssl verify -CAfile ../certs/ca.crt ../certs/client.crt
    ```
  - 测试证书
    - 连接到远程服务器 `openssl s_client -connect host.docker.internal:8000 -showcerts`
    - 带 CA 证书连接远程服务器 `openssl s_client -connect host.docker.internal:8000 -CAfile ca.crt`
    - 调试远程服务器的 SSL/TLS `openssl s_client -connect host.docker.internal:8000 -tlsextdebug`
    - 模拟的 HTTPS 服务，可以返回 Openssl 相关信息 `openssl s_server -accept 443 -cert server.crt -key server.key -www`
- [NAT](https://arthurchiao.art/blog/nat-zh/)
  - Netfilter
    - Linux 内核中有一个数据包过滤框架（packet filter framework），叫做 netfilter（ 项目地址 netfilter.org）。这个框架使得 Linux 机器可以像路由器一 样工作。
    - 和 NAT 相关的最重要的规则，都在 nat 这个（iptables）table 里。这个表有三个预置的 chain：PREROUTING, OUTPUT 和 POSTROUTING
  - 如何设置规则
    - 从本地网络发出的、目的是因特网的包，将发送方地址修改为路由器 的地址。 `iptables -t nat -A POSTROUTING -o eth1 -j MASQUERADE`
  - iptable
    ```shell
    $> iptables -t nat -A chain [...]
    
    # list rules:
    $> iptables -t nat -L
    
    # remove user-defined chain with index 'myindex':
    $> iptables -t nat -D chain myindex
    
    # Remove all rules in chain 'chain':
    $> iptables -t nat -F chain
    # TCP packets from 192.168.1.2:
    $> iptables -t nat -A POSTROUTING -p tcp -s 192.168.1.2 [...]
    
    # UDP packets to 192.168.1.2:
    $> iptables -t nat -A POSTROUTING -p udp -d 192.168.1.2 [...]
    
    # all packets from 192.168.x.x arriving at eth0:
    $> iptables -t nat -A PREROUTING -s 192.168.0.0/16 -i eth0 [...]
    
    # all packets except TCP packets and except packets from 192.168.1.2:
    $> iptables -t nat -A PREROUTING -p ! tcp -s ! 192.168.1.2 [...]
    
    # packets leaving at eth1:
    $> iptables -t nat -A POSTROUTING -o eth1 [...]
    
    # TCP packets from 192.168.1.2, port 12345 to 12356
    # to 123.123.123.123, Port 22
    # (a backslash indicates contination at the next line)
    $> iptables -t nat -A POSTROUTING -p tcp -s 192.168.1.2 \
       --sport 12345:12356 -d 123.123.123.123 --dport 22 [...]
    
    #对于 nat table，有如下几种动作：SNAT, MASQUERADE, DNAT, REDIRECT，都需要通过 -j 指定
    # Source-NAT: Change sender to 123.123.123.123
    $> iptables [...] -j SNAT --to-source 123.123.123.123
    
    # Mask: Change sender to outgoing network interface
    $> iptables [...] -j MASQUERADE
    
    # Destination-NAT: Change receipient to 123.123.123.123, port 22
    $> iptables [...] -j DNAT --to-destination 123.123.123.123:22
    
    # Redirect to local port 8080
    $> iptables [...] -j REDIRECT --to-ports 8080
    ```
    - SNAT - 修改源 IP 为固定新 IP 
    - MASQUERADE - 修改源 IP 为动态新 IP
    - DNAT - 修改目的 IP - DNAT 可以用于运行在防火墙后面的服务器。
    - REDIRECT - 将包重定向到本机另一个端口 - REDIRECT 是 DNAT 的一个特殊场景。包被重定向到路由器的另一个本地端口，可以实现， 例如透明代理的功能。和 DNAT 一样，REDIRECT 适用于 PREROUTING 和 OUTPUT chain 。
  - NAT 应用 
    - 透明代理 `iptables -t nat -A PREROUTING -i eth0 -p tcp --dport 80 -j REDIRECT --to-ports 8080 `
    - 绕过防火墙 
    - 通过 NAT 从外网访问内网服务 
      - 假设我 们有一个 HTTP 服务运行在内网机器 192.168.1.2，NAT 路由器的地址是 192.168.1.1 ，并通过另一张有公网 IP 123.123.123.123 的网卡连接到了外部网络。 要使得外网机器可以访问 192.168.1.2 的服务
      - `iptables -t nat -A PREROUTING -p tcp -i eth1 --dport 80 -j DNAT --to 192.168.1.2`
- [NAT 穿透是如何工作的](https://arthurchiao.art/blog/how-nat-traversal-works-zh/#77-%E5%85%A8-ipv6-%E7%BD%91%E7%BB%9C%E7%90%86%E6%83%B3%E4%B9%8B%E5%9C%B0%E4%BD%86%E5%B9%B6%E9%9D%9E%E9%97%AE%E9%A2%98%E5%85%A8%E6%97%A0)
  - NAT 设备是一个增强版的有状态防火墙
  - SNAT 的意义：解决 IPv4 地址短缺问题 - 将很多设备连接到公网，而只使用少数几个公网 IP
  - 穿透 “NAT+防火墙”：STUN (Session Traversal Utilities for NAT) 协议
    - STUN 基于一个简单的观察：从一个会被 NAT 的客户端访问公网服务器时， 服务器看到的是 NAT 设备的公网 ip:port 地址，而非该 客户端的局域网 ip:port 地址。
- [nslookup-OK-but-ping-fail](https://plantegg.github.io/2019/01/09/nslookup-OK-but-ping-fail/)
  - 域名解析流程
    - DNS域名解析的时候先根据 /etc/nsswitch.conf 配置的顺序进行dns解析（name service switch），一般是这样配置：hosts: files dns 【files代表 /etc/hosts ； dns 代表 /etc/resolv.conf】(ping是这个流程，但是nslookup和dig不是)
    - 如果本地有DNS Client Cache，先走Cache查询，所以有时候看不到DNS网络包。Linux下nscd可以做这个cache，Windows下有 ipconfig /displaydns ipconfig /flushdns
    - 如果 /etc/resolv.conf 中配置了多个nameserver，默认使用第一个，只有第一个失败【如53端口不响应、查不到域名后再用后面的nameserver顶上】
    - 如果 /etc/resolv.conf 中配置了rotate，那么多个nameserver轮流使用. 但是因为glibc库的原因用了rotate 会触发nameserver排序的时候第二个总是排在第一位
- [理解 netfilter 和 iptables](https://mp.weixin.qq.com/s/e8pyVJZ4CBf0xy3OGVr3LA)
  - Netfilter 的设计与实现
    - ![img.png](k8s_network_packet_path.png)
    - netfilter hooks
      - 所谓的 hook 实质上是代码中的枚举对象（值为从 0 开始递增的整型）每个 hook 在内核网络栈中对应特定的触发点位置，以 IPv4 协议栈为例，有以下 netfilter hooks 定义：
      - 所有的触发点位置统一调用 NF_HOOK 这个宏来触发 hook
    - 回调函数与优先级
      - netfilter 的另一组成部分是 hook 的回调函数。内核网络栈既使用 hook 来代表特定触发位置，也使用 hook （的整数值）作为数据索引来访问触发点对应的回调函数。
      - 内核的其他模块可以通过 netfilter 提供的 api 向指定的 hook 注册回调函数，同一 hook 可以注册多个回调函数，通过注册时指定的 priority 参数可指定回调函数在执行时的优先级。
  - iptables
    - 基于内核 netfilter 提供的 hook 回调函数机制，netfilter 作者 Rusty Russell 还开发了 iptables，实现在用户空间管理应用于数据包的自定义规则。
      - 用户空间的 iptables 命令向用户提供访问内核 iptables 模块的管理界面。
      - 内核空间的 iptables 模块在内存中维护规则表，实现表的创建及注册。
      - iptables 主要操作以下几种对象：
        - table：对应内核空间的 xt_table 结构，iptable 的所有操作都对指定的 table 执行，默认为 filter。
        - chain：对应指定 table 通过特定 netfilter hook 调用的规则集，此外还可以自定义规则集，然后从 hook 规则集中跳转过去。
        - rule：对应上文中 ipt_entry、ipt_entry_match 和 ipt_entry_target，定义了对数据包的匹配规则以及匹配后执行的行为。
        - match：具有很强扩展性的自定义匹配规则。
        - target：具有很强扩展性的自定义匹配后行为。
  - conntrack
    - 仅仅通过 3、4 层的首部信息对数据包进行过滤是不够的，有时候还需要进一步考虑连接的状态。netfilter 通过另一内置模块 conntrack 进行连接跟踪（connection tracking），以提供根据连接过滤、地址转换（NAT）等更进阶的网络过滤功能。由于需要对连接状态进行判断，conntrack 在整体机制相同的基础上，又针对协议特点有单独的实现。







