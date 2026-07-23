# Kubernetes service account

Chaincode pods are created with a service account defined by the `FABRIC_K8S_BUILDER_SERVICE_ACCOUNT` environment variable, or the `default` service account if the variable is not set.

If your chaincode images are published to registries which require credentials, you will need to add image pull secrets to the service account.

For example, follow these steps if `FABRIC_K8S_BUILDER_NAMESPACE` and `FABRIC_K8S_BUILDER_SERVICE_ACCOUNT` are both set to `hlf-chaincode`.

Create the `hlf-chaincode` service account.

```shell
kubectl create serviceaccount hlf-chaincode --namespace=hlf-chaincode
```

Create an imagePullSecret.

```shell
kubectl create secret docker-registry hlf-fabregistry-key --namespace=hlf-chaincode \
    --docker-server=DOCKER_SERVER \
    --docker-username=DOCKER_USERNAME \
    --docker-password=DOCKER_PASSWORD \
    --docker-email=DOCKER_EMAIL
```

Add the image pull secret to the service account.

```shell
kubectl patch serviceaccount hlf-chaincode --namespace=hlf-chaincode \
    -p '{"imagePullSecrets": [{"name": "hlf-fabregistry-key"}]}'
```

See the Kubernetes [Configure Service Accounts for Pods](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#add-imagepullsecrets-to-a-service-account) documentation for details.

## Alternative: using FABRIC_K8S_BUILDER_IMAGE_PULL_SECRETS

As an alternative to patching the service account, you can attach image pull secrets directly to the chaincode pod using the `FABRIC_K8S_BUILDER_IMAGE_PULL_SECRETS` environment variable. This is useful when you want to use different pull secrets for different chaincodes, or when you don't want to modify the service account.

Set the variable to a JSON array of secret names on the peer:

```shell
export FABRIC_K8S_BUILDER_IMAGE_PULL_SECRETS='["hlf-fabregistry-key"]'
```

And add it to `propagateEnvironment` in `core.yaml`:

```yaml
externalBuilders:
  - name: k8s_builder
    path: /opt/hyperledger/k8s_builder
    propagateEnvironment:
      - FABRIC_K8S_BUILDER_IMAGE_PULL_SECRETS
```

The secrets must already exist in the chaincode namespace before the chaincode pod starts. Both approaches (service account patching and this variable) can be used together.
