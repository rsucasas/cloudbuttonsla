# Prometheus, NFS 

-----------------------


### Prometheus

-----------------------

#### NFS, Volumes

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: nfs
spec:
  capacity:
    storage: 3000Gi
  accessModes:
    - ReadWriteMany
  nfs:
    server: 77.231.202.1
    path: "/db/"
```

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: nfs
spec:
  accessModes:
    - ReadWriteMany
  storageClassName: ""
  resources:
    requests:
      storage: 3000Gi
```

-----------------------

#### Deployment.yaml

**serviceAccount**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    deployment.kubernetes.io/revision: "3"
  generation: 1
  labels:
    app: prometheus-nfs
    project: prometheus-nfs
  name: prometheus-nfs
spec:
  progressDeadlineSeconds: 600
  replicas: 1
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: prometheus-nfs
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: prometheus-nfs
    spec:
      containers:
      - image: prom/prometheus
        imagePullPolicy: IfNotPresent
        name: prometheus-nfs
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        args:
          - '--config.file=/etc/prometheus/prometheus.yaml'
          - '--storage.tsdb.retention=365d'
          - '--storage.tsdb.no-lockfile'
        ports:
          - containerPort: 9090
        volumeMounts:
          - name: prometheus-config-volume
            mountPath: /etc/prometheus/
          - name: prometheus-storage-volume
            mountPath: prometheus/
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      serviceAccountName: prometheus-system
      serviceAccount: prometheus-system
      securityContext: {}
      terminationGracePeriodSeconds: 30
      initContainers:
      - name: prometheus-data-permission-fix
        image: busybox
        command: ["/bin/chmod","-R","777", "/prometheus"]
        volumeMounts:
        - name: prometheus-storage-volume
          mountPath: /prometheus
      volumes:
        - name: prometheus-config-volume
          configMap:
            name: prometheus-configuration-p-mm6gb52655
        - name: prometheus-storage-volume
          persistentVolumeClaim:
            claimName: nfs
status: {}
```

-----------------------

#### Service.yaml

```yaml
kind: Service
apiVersion: v1
metadata:
  name: prometheus-nfs-service
  namespace: knative-monitoring
spec:
  ports:
    - protocol: TCP
      port: 8080
      targetPort: 9090
      nodePort: 30990
  selector:
    app: prometheus-nfs
  clusterIP: 10.99.17.5
  type: NodePort
  sessionAffinity: None
  externalTrafficPolicy: Cluster
status:
  loadBalancer: {}
```

-----------------------

#### Prometheus.yaml

```yaml
global:
    scrape_interval: 30s
    scrape_timeout: 10s
    evaluation_interval: 30s
    
scrape_configs: 

# Controller endpoint 

- job_name: controller
    scrape_interval: 3s
    scrape_timeout: 3s
    kubernetes_sd_configs:
    - role: pod
    relabel_configs:
    # Scrape only the the targets matching the following metadata
    - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_pod_label_app, __meta_kubernetes_pod_container_port_name]
    action: keep
    regex: knative-serving;controller;metrics
    # Rename metadata labels to be reader friendly
    - source_labels: [__meta_kubernetes_namespace]
    target_label: namespace
    - source_labels: [__meta_kubernetes_pod_name]
    target_label: pod
    - source_labels: [__meta_kubernetes_service_name]
    target_label: service
    
# Pushgateway 

- job_name: pushgateway
    scrape_interval: 3s
    scrape_timeout: 3s  
    kubernetes_sd_configs:
    - role: pod
    relabel_configs:
    # Scrape only the targets matching the following metadata
    - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_pod_label_app, __meta_kubernetes_pod_container_port_name]
    action: keep
    regex: knative-monitoring;prometheus-pushgateway;metrics
    # Rename metadata labels to be reader friendly
    - source_labels: [__meta_kubernetes_namespace]
    target_label: namespace
    - source_labels: [__meta_kubernetes_pod_name]
    target_label: pod
    - source_labels: [__meta_kubernetes_service_name]
    target_label: service   

# Autoscaler endpoint 

- job_name: autoscaler
    scrape_interval: 3s
    scrape_timeout: 3s
    kubernetes_sd_configs:
    - role: pod
    relabel_configs:
    # Scrape only the the targets matching the following metadata
    - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_pod_label_app, __meta_kubernetes_pod_container_port_name]
    action: keep
    regex: knative-serving;autoscaler;metrics
    # Rename metadata labels to be reader friendly
    - source_labels: [__meta_kubernetes_namespace]
    target_label: namespace
    - source_labels: [__meta_kubernetes_pod_name]
    target_label: pod
    - source_labels: [__meta_kubernetes_service_name]
    target_label: service
    
# Activator pods 

- job_name: activator
    scrape_interval: 3s
    scrape_timeout: 3s
    kubernetes_sd_configs:
    - role: pod
    relabel_configs:
    # Scrape only the the targets matching the following metadata
    - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_pod_label_app, __meta_kubernetes_pod_container_port_name]
    action: keep
    regex: knative-serving;activator;metrics
    # Rename metadata labels to be reader friendly
    - source_labels: [__meta_kubernetes_namespace]
    target_label: namespace
    - source_labels: [__meta_kubernetes_pod_name]
    target_label: pod
    - source_labels: [__meta_kubernetes_service_name]
    target_label: service
    
# Webhook pods 

- job_name: webhook
    scrape_interval: 3s
    scrape_timeout: 3s
    kubernetes_sd_configs:
    - role: pod
    relabel_configs:
    # Scrape only the the targets matching the following metadata
    - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_pod_label_role, __meta_kubernetes_pod_container_port_name]
    action: keep
    regex: knative-serving;webhook;metrics
    # Rename metadata labels to be reader friendly
    - source_labels: [__meta_kubernetes_namespace]
    target_label: namespace
    - source_labels: [__meta_kubernetes_pod_name]
    target_label: pod
    - source_labels: [__meta_kubernetes_service_name]
    target_label: service
    
# Queue proxy metrics 

- job_name: queue-proxy
    scrape_interval: 3s
    scrape_timeout: 3s
    kubernetes_sd_configs:
    - role: pod
    relabel_configs:
    # Scrape only the the targets matching the following metadata
    - source_labels: [__meta_kubernetes_pod_label_serving_knative_dev_revision, __meta_kubernetes_pod_container_port_name]
    action: keep
    regex: .+;http-usermetric
    # Rename metadata labels to be reader friendly
    - source_labels: [__meta_kubernetes_namespace]
    target_label: namespace
    - source_labels: [__meta_kubernetes_pod_name]
    target_label: pod
    - source_labels: [__meta_kubernetes_service_name]
    target_label: service
    
# Fluentd daemonset 

- job_name: fluentd-ds
    kubernetes_sd_configs:
    - role: endpoints
    relabel_configs:
    # Scrape only the the targets matching the following metadata
    - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_label_app, __meta_kubernetes_endpoint_port_name]
    action: keep
    regex: knative-monitoring;fluentd-ds;prometheus-metrics
    # Rename metadata labels to be reader friendly
    - source_labels: [__meta_kubernetes_namespace]
    target_label: namespace
    - source_labels: [__meta_kubernetes_pod_name]
    target_label: pod
    - source_labels: [__meta_kubernetes_service_name]
    target_label: service
    
# Istio mesh 

- job_name: istio-mesh
    scrape_interval: 5s
    kubernetes_sd_configs:
    - role: endpoints
    relabel_configs:
    # Scrape only the the targets matching the following metadata
    - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
    action: keep
    regex: istio-system;istio-telemetry;prometheus
    # Rename metadata labels to be reader friendly
    - source_labels: [__meta_kubernetes_namespace]
    target_label: namespace
    - source_labels: [__meta_kubernetes_pod_name]
    target_label: pod
    - source_labels: [__meta_kubernetes_service_name]
    target_label: service
    
# Istio policy 

- job_name: istio-policy
    scrape_interval: 5s
    kubernetes_sd_configs:
    - role: endpoints
    relabel_configs:
    # Scrape only the the targets matching the following metadata
    - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
    action: keep
    regex: istio-system;istio-policy;http-monitoring
    # Rename metadata labels to be reader friendly
    - source_labels: [__meta_kubernetes_namespace]
    target_label: namespace
    - source_labels: [__meta_kubernetes_pod_name]
    target_label: pod
    - source_labels: [__meta_kubernetes_service_name]
    target_label: service
    
# Istio telemetry 

- job_name: istio-telemetry
    scrape_interval: 5s
    kubernetes_sd_configs:
    - role: endpoints
    relabel_configs:
    # Scrape only the the targets matching the following metadata
    - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
    action: keep
    regex: istio-system;istio-telemetry;http-monitoring
    # Rename metadata labels to be reader friendly
    - source_labels: [__meta_kubernetes_namespace]
    target_label: namespace
    - source_labels: [__meta_kubernetes_pod_name]
    target_label: pod
    - source_labels: [__meta_kubernetes_service_name]
    target_label: service
    
# Istio pilot 

- job_name: istio-pilot
    scrape_interval: 5s
    kubernetes_sd_configs:
    - role: endpoints
    relabel_configs:
    # Scrape only the the targets matching the following metadata
    - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
    action: keep
    regex: istio-system;istio-pilot;http-monitoring
    # Rename metadata labels to be reader friendly
    - source_labels: [__meta_kubernetes_namespace]
    target_label: namespace
    - source_labels: [__meta_kubernetes_pod_name]
    target_label: pod
    - source_labels: [__meta_kubernetes_service_name]
    target_label: service
    
# Kube API server 

- job_name: kube-apiserver
    scheme: https
    kubernetes_sd_configs:
    - role: endpoints
    bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
    tls_config:
    ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    server_name: kubernetes
    insecure_skip_verify: false
    relabel_configs:
    # Scrape only the the targets matching the following metadata
    - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_label_component, __meta_kubernetes_service_label_provider, __meta_kubernetes_endpoint_port_name]
    action: keep
    regex: default;apiserver;kubernetes;https
    # Rename metadata labels to be reader friendly
    - source_labels: [__meta_kubernetes_namespace]
    target_label: namespace
    - source_labels: [__meta_kubernetes_pod_name]
    target_label: pod
    - source_labels: [__meta_kubernetes_service_name]
    target_label: service
    
# Kube controller manager 

- job_name: kube-controller-manager
    kubernetes_sd_configs:
    - role: endpoints
    relabel_configs:
    # Scrape only the the targets matching the following metadata
    - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_label_app, __meta_kubernetes_endpoint_port_name]
    action: keep
    regex: knative-monitoring;kube-controller-manager;http-metrics
    # Rename metadata labels to be reader friendly
    - source_labels: [__meta_kubernetes_namespace]
    target_label: namespace
    - source_labels: [__meta_kubernetes_pod_name]
    target_label: pod
    - source_labels: [__meta_kubernetes_service_name]
    target_label: service
    
# Kube scheduler 

- job_name: kube-scheduler
    kubernetes_sd_configs:
    - role: endpoints
    relabel_configs:
    # Scrape only the the targets matching the following metadata
    - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_label_k8s_app, __meta_kubernetes_endpoint_port_name]
    action: keep
    regex: kube-system;kube-scheduler;http-metrics
    # Rename metadata labels to be reader friendly
    - source_labels: [__meta_kubernetes_namespace]
    target_label: namespace
    - source_labels: [__meta_kubernetes_pod_name]
    target_label: pod
    - source_labels: [__meta_kubernetes_service_name]
    target_label: service
    
# Kube state metrics on https-main port 

- job_name: kube-state-metrics-http-metrics
    honor_labels: true
    kubernetes_sd_configs:
    - role: endpoints
    relabel_configs:
    # Scrape only the the targets matching the following metadata
    - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_label_app, __meta_kubernetes_endpoint_port_name]
    action: keep
    regex: knative-monitoring;kube-state-metrics;http-metrics
    # Rename metadata labels to be reader friendly
    - source_labels: [__meta_kubernetes_namespace]
    target_label: namespace
    - source_labels: [__meta_kubernetes_pod_name]
    target_label: pod
    - source_labels: [__meta_kubernetes_service_name]
    target_label: service
    
# Kube state metrics on https-self port 

- job_name: kube-state-metrics-telemetry
    kubernetes_sd_configs:
    - role: endpoints
    relabel_configs:
    # Scrape only the the targets matching the following metadata
    - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_label_app, __meta_kubernetes_endpoint_port_name]
    action: keep
    regex: knative-monitoring;kube-state-metrics;telemetry
    # Rename metadata labels to be reader friendly
    - source_labels: [__meta_kubernetes_namespace]
    target_label: namespace
    - source_labels: [__meta_kubernetes_pod_name]
    target_label: pod
    - source_labels: [__meta_kubernetes_service_name]
    target_label: service
    
# Kubelet - nodes

- job_name: kubernetes-nodes
    scheme: https
    tls_config:
    ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
    kubernetes_sd_configs:
    - role: node
    relabel_configs:
    - action: labelmap
    regex: __meta_kubernetes_node_label_(.+)
    - target_label: __address__
    replacement: kubernetes.default.svc:443
    - source_labels: [__meta_kubernetes_node_name]
    target_label: __metrics_path__
    replacement: /api/v1/nodes/${1}/proxy/metrics
    
# Kubelet - cAdvisor 

- job_name: kubernetes-cadvisor
    scrape_interval: 15s
    scheme: https
    tls_config:
    ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
    kubernetes_sd_configs:
    - role: node
    relabel_configs:
    - action: labelmap
    regex: __meta_kubernetes_node_label_(.+)
    - target_label: __address__
    replacement: kubernetes.default.svc:443
    - source_labels: [__meta_kubernetes_node_name]
    target_label: __metrics_path__
    replacement: /api/v1/nodes/${1}/proxy/metrics/cadvisor
    
# Node exporter 

- job_name: node-exporter
    scheme: https
    kubernetes_sd_configs:
    - role: endpoints
    bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
    tls_config:
    insecure_skip_verify: true
    relabel_configs:
    # Scrape only the the targets matching the following metadata
    - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_label_app, __meta_kubernetes_endpoint_port_name]
    action: keep
    regex: knative-monitoring;node-exporter;https
    # Rename metadata labels to be reader friendly
    - source_labels: [__meta_kubernetes_namespace]
    target_label: namespace
    - source_labels: [__meta_kubernetes_pod_name]
    target_label: pod
    - source_labels: [__meta_kubernetes_service_name]
    target_label: service
    
# Prometheus 

- job_name: prometheus
    kubernetes_sd_configs:
    - role: endpoints
    relabel_configs:
    # Scrape only the the targets matching the following metadata
    - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_label_app, __meta_kubernetes_endpoint_port_name]
    action: keep
    regex: knative-monitoring;prometheus;web
    # Rename metadata labels to be reader friendly
    - source_labels: [__meta_kubernetes_namespace]
    target_label: namespace
    - source_labels: [__meta_kubernetes_pod_name]
    target_label: pod
    - source_labels: [__meta_kubernetes_service_name]
    target_label: service
    
- job_name: kubernetes-service-endpoints
    kubernetes_sd_configs:
    - role: endpoints
    relabel_configs:
    - action: keep
    regex: true
    source_labels:
    - __meta_kubernetes_service_annotation_prometheus_io_scrape
    - action: replace
    regex: ([^:]+)(?::\d+)?;(\d+)
    replacement: $1:$2
    source_labels:
    - __address__
    - __meta_kubernetes_service_annotation_prometheus_io_port
    target_label: __address__
    - action: labelmap
    regex: __meta_kubernetes_service_label_(.+)
    - action: replace
    source_labels:
    - __meta_kubernetes_namespace
    target_label: kubernetes_namespace
    - action: replace
    source_labels:
    - __meta_kubernetes_service_name
    target_label: kubernetes_name     
```