# Deployment instructions

(All instructions are intended to be run from this folder)

```
docker-compose up -d
```

Use docker images to see the created image

```
$ docker images
REPOSITORY                           TAG                 IMAGE ID            CREATED             SIZE
<docker-image>                       latest              119c15ed2be9        19 hours ago        21.8MB
alpine                               3.8                 3f53bb00af94        3 weeks ago         4.41MB
golang                               1.10.3-alpine3.8    cace225819dc        4 months ago        259MB
```

*\<docker-image> is used in place of the image name defined in docker-compose*


## Log in to the container registry

To use the ACR (Azure Container Repository) instance, you must first log in. Use the az acr login command and provide the unique name of your acr

```
az acr login --name <acrName>
```

Get the login server address for the acr by replacing \<resource-group> with the resource group name

```
az acr list --resource-group <resource-group> --query "[].{acrLoginServer:loginServer}" --output table
```

Now, tag your local docker-image with the acrloginServer address of the container registry. To indicate the image version, add a version at the end of the tag

```
docker tag <docker-image> <acrLoginServer>/<docker-image>:<version-number>
```

To verify the tags are applied, run docker images again. An image is tagged with the ACR instance address and a version number.

```
$ docker images
REPOSITORY                           TAG                 IMAGE ID            CREATED             SIZE
<docker-image>                       latest              119c15ed2be9        19 hours ago        21.8MB
<acrname>.azurecr.io/<docker-image>  <version-number>    a6981eabaf5f        20 hours ago        21.8MB
alpine                               3.8                 3f53bb00af94        3 weeks ago         4.41MB
golang                               1.10.3-alpine3.8    cace225819dc        4 months ago        259MB
```

## Push image to registry
With your image built and tagged, push the docker-image to your ACR instance. Use docker push and provide your own acrLoginServer address for the image name

```
docker push <acrLoginServer>/<docker-image>:<version-number>
```

## Create Kubernetes AzureIdentity and AzureIdentityBinding

To be able to bind the managed identity you’ve created to the pod that will run the sample application, you need to define two new Kubernetes resources: an AzureIdentity and an AzureIdentityBinding.

```
kubectl apply -f azureidentity.yaml
```

```
kubectl apply -f azureidentitybinding.yaml
```

## Deploy the application

Use kubectl apply create the defined Kubernetes objects

```
kubectl apply -f keyvaultsample.yaml
```

## Check deployment

To monitor progress, use the kubectl get service command with the --watch argument.

```
$ kubectl get service <app-name> --watch
NAME             TYPE           CLUSTER-IP    EXTERNAL-IP   PORT(S)        AGE
<app-name>       LoadBalancer   10.0.133.91   <pending>     80:30909/TCP   43s
```

When the EXTERNAL-IP address changes from pending to an actual public IP address, use CTRL-C to stop the kubectl watch process. The following example output shows a valid public IP address assigned to the service

```
NAME             TYPE           CLUSTER-IP    EXTERNAL-IP    PORT(S)        AGE
<app-name>       LoadBalancer   10.0.133.91   13.68.225.93   80:30909/TCP   2m
```