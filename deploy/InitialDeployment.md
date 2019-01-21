# Configure Initial Azure Deployment

This tutorial requires that you're running the Azure CLI. Run `az --version` to find the version. If you need to install or upgrade, see [here](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest).

## Create an Azure Keyvault instance

*Skip this step if attempting to use aad-pod-identity and managed identities to access any other resource. Just replace the keyvault with that resource from here*

You can create an Azure Keyvault instance using the following commands. Replace `<keyvaultResourceGroup>` with the name of the resource group containing the keyvault and `<keyvaultName>` with the name of the keyvault.

```
az group create -n <keyvaultResourceGroup> -l eastus
az keyvault create -n <keyvaultName> -g <keyvaultResourceGroup> -l eastus
```

Once the Azure Keyvault instance is ready, you can create a secret:

```
az keyvault secret set -n mySecret --vault-name <keyvaultName> --value MySuperSecretThatIDontWantToShareWithYou!
```

Now, this project is going to retrieve the value of the secret `mySecret` from inside a pod running in the Kubernetes cluster, without passing any credentials to the application! But first thing first, let’s setup your AKS cluster!


## Create an Azure Container Registry

To create an Azure Container Registry, you first need a resource group. An Azure resource group is a logical container into which Azure resources are deployed and managed.

Create a resource group. In the following example, a resource group is created in the eastus region. Replace `<resourceGroupName>` with the name of your choice.

```
az group create --name <resourceGroupName> --location eastus
```

Create an Azure Container Registry instance. The registry name must be unique within Azure, and contain 5-50 alphanumeric characters. Replace `<acrName>` with the name of your choice.

```
az acr create --resource-group <resourceGroupName> --name <acrName> --sku Basic
```

## Create a service principal

To allow an AKS cluster to interact with other Azure resources, an Azure Active Directory service principal is used. This service principal can be automatically created by the Azure CLI or portal, or you can pre-create one and assign additional permissions.

The following creates a service principal. The --skip-assignment parameter limits any additional permissions from being assigned. By default, this service principal is valid for one year.

```
az ad sp create-for-rbac --skip-assignment
```

The output is similar to the following example:

```
{
  "appId": "e7596ae3-6864-4cb8-94fc-20164b1588a9",
  "displayName": "azure-cli-2018-06-29-19-14-37",
  "name": "http://azure-cli-2018-06-29-19-14-37",
  "password": "52c95f25-bd1e-4314-bd31-d8112b293521",
  "tenant": "72f988bf-86f1-41af-91ab-2d7cd011db48"
}
```

Make a note of the `appId` and `password`. These values are used in the following steps. The `appId` here is the `<servicePrincipleId>` that will be used later.

## Configure ACR authentication

To access images stored in ACR, you must grant the AKS service principal the correct rights to pull images from ACR.

First, get the ACR resource ID

```
az acr show --resource-group <resourceGroupName> --name <acrName> --query "id" --output tsv
```

To grant the correct access for the AKS cluster to use images stored in ACR, create a role assignment using the az role assignment create command. Replace `<servicePrincipleId>` and `<acrId>` with the values gathered in the previous two steps.

```
az role assignment create --assignee <servicePrincipleId> --scope <acrId> --role Reader
```

## Create Kubernetes Cluster

AKS clusters can use Kubernetes role-based access controls (RBAC). These controls let you define access to resources based on roles assigned to users. Permissions are combined if a user is assigned multiple roles, and permissions can be scoped to either a single namespace or across the whole cluster. By default, the Azure CLI automatically enables RBAC when you create an AKS cluster.

Create an AKS cluster using az aks create. The following example creates a cluster  in the resource group created previously. Provide the `<appId>` and `<password>` from when the service principal was created. Replace `<clusterName>` with the name of your cluster.

```
az aks create --resource-group <resourceGroupName> --name <clusterName> --node-count 1 --service-principal <appId> --client-secret <password> --generate-ssh-keys
```

## Install the Kubernetes CLI

To connect to the Kubernetes cluster, you use [kubectl](https://kubernetes.io/docs/user-guide/kubectl/), the Kubernetes command-line client.

If you use the Azure Cloud Shell, `kubectl` is already installed. You can also install it locally using the az aks install-cli command

```
az aks install-cli
```

## Connect to cluster using kubectl

Configure `kubectl` to connect to your Kubernetes cluster

```
az aks get-credentials --resource-group <resourceGroupName> --name <clusterName>
```

## Configure your Kubernetes cluster to run Azure AD Pod Identity infrastructure

If you want the full details, everything to know is well documented on the [project page](https://github.com/Azure/aad-pod-identity#get-started).

As RBAC is automatically enabled on AKS, install the RBAC deployment:

```
kubectl create -f https://raw.githubusercontent.com/Azure/aad-pod-identity/master/deploy/infra/deployment-rbac.yaml
```

## Create an Azure managed identity

Now that your Kubernetes cluster is ready to provide Azure Active Directory tokens to your applications, you need to create an Azure Managed Identity and assign role to it. This is the identity that you will later bind on your pod running the sample application.

To create a managed identity, you can use this command

```
az identity create -n keyvaultsampleidentity -g <keyvaultResourceGroup>
```

**Note: keep the `principalId` and `clientId` from the output of this command, you will need it later.**

Then you need to make sure the managed identity has `Reader` role on the Azure KeyVault resource:

```
az role assignment create --role "Reader" --assignee <principalId> --scope /subscriptions/{YourSubscriptionID}/resourceGroups/<keyvaultResourceGroup>/providers/Microsoft.KeyVault/vaults/<keyvaultName>
```

And that it can access the secrets

```
az keyvault set-policy -n <keyvaultName> --secret-permissions get list --spn <clientid>
```

[aad-pod-identity uses the service principal of your Kubernetes cluster](https://github.com/Azure/aad-pod-identity#providing-required-permissions-for-mic) to access the Azure managed identity resource and work with it.
This is why you need to give this service principal the rights to use the managed identity created before

```
az role assignment create --role "Managed Identity Operator" --assignee <servicePrincipleId> --scope /subscriptions/{YourSubscriptionID}/resourceGroups/<keyvaultResourceGroup>/providers/Microsoft.ManagedIdentity/userAssignedIdentities/keyvaultsampleidentity
```

*Note: The `<servicePrincipalId>` is the `appId` of the service principal created earlier.*

If you have lost the `<servicePrincipalId>`, it can be found by running

```
az aks show --resource-group <resourceGroupName> --name <clusterName>
```

## Create Kubernetes AzureIdentity and AzureIdentityBinding

To be able to bind the managed identity you’ve created to the pod that will run the sample application, you need to define two new Kubernetes resources: an AzureIdentity and an AzureIdentityBinding.

Update azureidentity.yaml

```
apiVersion: "aadpodidentity.k8s.io/v1"
kind: AzureIdentity
metadata:
  name: keyvaultsampleidentity
spec:
  type: 0
  ResourceID: /subscriptions/{YourSubscriptionID}/resourceGroups/<keyvaultResourceGroup>/providers/Microsoft.ManagedIdentity/userAssignedIdentities/keyvaultsampleidentity
  ClientID: <clientid>
```

*Note: you will find the `clientid` field the output of the identity creation command.*

```
kubectl apply -f azureidentity.yaml
```

Then update azureidentitybinding.yaml

```
azureidentitybinding.yaml:
apiVersion: "aadpodidentity.k8s.io/v1"
kind: AzureIdentityBinding
metadata:
  name: keyvaultsampleidentity-binding
spec:
  AzureIdentity: keyvaultsampleidentity
  Selector: keyvaultsampleidentity
```

*Note: the value of the `Selector` property in the YAML definition above will be used to bind the Azure identity to your pod, using labels in its specifications.*

```
kubectl apply -f azureidentitybinding.yaml
```

## Deploy the sample application

Follow the [deployment instructions](./README.md)