# InfluxDB 


-----------------------

#### Secrets

```bash
kubectl create secret generic influxdb-creds \
  --from-literal=INFLUXDB_DATABASE=influxdb \
  --from-literal=INFLUXDB_USERNAME=root \
  --from-literal=INFLUXDB_PASSWORD=root \
  --from-literal=INFLUXDB_HOST=influxdb
```

#### PV.yaml

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

#### PVC.yaml

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

#### Deployment.yaml

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    deployment.kubernetes.io/revision: "3"
  creationTimestamp: null
  generation: 1
  labels:
    app: influxdb
    project: influxdb
  name: influxdb
  selfLink: /apis/extensions/v1beta1/namespaces/default/deployments/influxdb
spec:
  progressDeadlineSeconds: 600
  replicas: 1
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: influxdb
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: influxdb
    spec:
      containers:
      - envFrom:
        - secretRef:
            name: influxdb-creds
        image: docker.io/influxdb:1.6.4
        imagePullPolicy: IfNotPresent
        name: influxdb
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /var/lib/influxdb
          name: var-lib-influxdb
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      terminationGracePeriodSeconds: 30
      volumes:
      - name: var-lib-influxdb
        persistentVolumeClaim:
          claimName: nfs
status: {}
```

#### Service.yaml

```yaml
apiVersion: v1
kind: Service
metadata:
  name: influxdb
  namespace: default
  labels:
    app: influxdb
    project: influxdb
spec:
  type: NodePort
  ports:
    - protocol: TCP
      port: 8086
      targetPort: 8086
      nodePort: 30086
  sessionAffinity: None
  selector:
    app: influxdb
```  

