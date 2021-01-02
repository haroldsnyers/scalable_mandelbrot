# Servers

This a system made of a proxy and different servers that run in docker containers. The different containers that will be 
initialised are defined in the docker-compose.yml file. In this file we define a proxy (application on which the servers 
sign up ) and which can also be accessed to retrieve the different servers that are running. This system enables us to 
scale our application mandelbrot horizontally as we can add more servers easily. 

Contrary to the application running on Docker, this proxy has one more task, which is to redirect the generation get 
methods from the client to the slaves. This way the client which will ask for the mandelbrot image will have no clue how 
many servers are running or at least only needs to know the proxy's address which will redirect to the intended server / 
slave. The proxy will thus act as a middle man to redirect the get request from the client and then sent the result back
from the slaves in the intended order. 

Initially, it was not intended that the client application had to generate the final image by itself and pass by a 
generator that would handle everything and at the end only send the bytes of the final image to decode. However, a problem
seems to occur when you try to make an http request to a server that in turn will execute http requests asynchronously to
finally generate the image and send the final image. The connection to the different servers from the generator was always
interrupted. 

This system runs in a kubernetes cluster. To test it, I used a minikube to test the cluster locally. I run the minikube 
application on wsl2 on windows. Here under are the steps you need to do to replicate the application. 
## Server configuration
First, you will need to install minikube on your linux (but can also work on windows). Then you can start the minikube 
application with `start minikube`. When minikube is running, type the following:

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
To see your pods running, you can type `kubectl get pods`.

A nice tool is octant, which lets you interact with your cluster in a more user-friendly way. 

Finally, you can execute the following command too to expose the proxy to outside world (outside the kubernetes cluster).

```shell script
minikube service proxy
```

The url that is returned will then need to be replaced inside the client.go file in the variable proxyServer. Then you 
run the client which will generate the mandelbrot. Again you can then play with the escape and width variable to increase
the depths and the quality of the mandelbrot respectively. 