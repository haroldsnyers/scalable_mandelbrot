# Servers

This a system made of a proxy and different servers that run in docker containers. The different containers that will be initialised are defined in the docker-compose.yml file. In this file we define a proxy (application on which the servers sign up ) and which can also be accessed to retrieve the different servers that are running. This system enables us to scale our application mandelbrot horizontally as we can add more servers easily. 

## Server configuration
To run the docker server configuration you will simply need to open a command line and go the "servers" folder.
At thi spoint first you will need to create if it is not already done a network. This can be done by typing the following.

```
docker network create resolute
```

Next you will be able to run the docker-compose which will launch the proxy-like application and the different servers as well. This can be done with the following command : 
```shell script
docker-compose -f docker-compose.yml up
```
You can also add `-d` flag to run this command in the background.

Finally, you will be able to access the list of servers up by going to 
`localhost:8090/get_computation`

## Adding Servers

To add servers, multiple solutions exist. At the startup, you can simply add more applications such as to have the number of server images in docker-compose that you want. However, it is also possible to add a seperate server by building and running an image externally (manually)

### adding a server in the docker compose

### adding a server manually with docker build and run 