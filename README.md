# Azure AAD Pod Identity Keyvault Example

This add on requires the Azure AD Pod Identity infrastructure.

To install, use

```
kubectl create -f https://raw.githubusercontent.com/Azure/aad-pod-identity/master/deploy/infra/deployment-rbac.yaml
```

## Building locally

To build this project locally, ensure you have docker installed and configured correctly (instructions are [here](https://docs.docker.com/)).

Then run

```
docker-compose up
```

access the local deployment at http://localhost:8080

## Deployment

Deployment instructions are [here](./deploy/README.md)


## Resources

More information on how to configure managed identity access:
https://blog.jcorioland.io/archives/2018/09/05/azure-aks-active-directory-managed-identities.html


[AKS (Azure Kubernetes Service) docs](https://docs.microsoft.com/en-us/azure/aks/)
