# Servers

This a system made of a proxy and different servers that run in docker containers. The different containers that will be initialised are defined in the docker-compose.yml file. In this file we define a proxy (application on which the servers sign up ) and which can also be accessed to retrieve the different servers that are running. This system enables us to scale our application mandelbrot horizontally as we can add more servers easily. 

Contrary to the application runnning on docker, this application contains also a generator. This way the client which will ask for the mandelbrot image will have no clue how many servers are running (and doesn't care). It will connect to the generator and request the image. The generator will then ask which server is online and then generate the image which he will send back to the client. 

This system runs in a kubernetes cluster. To test it, I used a minkube to test the cluster locally. I run the minikube application on wsl2 on windows. Here under are the steps you need to do to replicate the application. 
## Server configuration
First, you will need to install minikube on your linux (but can also work on windows). Then you can start the minkube application with `start minikube`. When minikube is running, type the following:

```shell script
# linking your docker environment
eval $(minikube docker-env) 
```

Then, you will need to build your docker images to be able tu use them with kubernetes. To do so, go to `horizontal_k8s`in the command line and type the following : 

```shell script
docker-compose build
```

After you have build and thus tagged your images properly, you can proceed with the following command : 

```shell script
kubectl apply -f ./k8s
```

This will execute the kubernetes resources and start your deployments and services. 
TO see your pods running, you can type `kubectl get pods`.

A nice tool is octant, which lets you interact with your cluster in a more user-friendly way. 

