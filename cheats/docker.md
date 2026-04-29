---
description: Container management — build, run, and orchestrate containers
tags: [containers, devops, deployment]
---

# docker

## Images
    docker images                         # list local images
    docker pull nginx:latest              # pull image
    docker build -t myapp:1.0 .          # build from Dockerfile
    docker build --no-cache -t myapp .   # build without cache
    docker rmi myapp:1.0                  # remove image
    docker image prune                    # remove dangling images

## Containers
    docker run nginx                      # run container (foreground)
    docker run -d nginx                   # run detached
    docker run -it ubuntu bash            # interactive shell
    docker run --rm alpine echo hello     # run and auto-remove
    docker run -p 8080:80 nginx           # map host:container ports
    docker run -v /host/path:/container/path nginx  # bind mount
    docker run -e ENV_VAR=value nginx     # set environment variable
    docker run --name mycontainer nginx   # named container

## Container Management
    docker ps                             # list running containers
    docker ps -a                          # list all containers
    docker stop <id>                      # graceful stop (SIGTERM)
    docker kill <id>                      # force kill (SIGKILL)
    docker rm <id>                        # remove stopped container
    docker rm -f <id>                     # force remove running container
    docker restart <id>                   # restart container

## Inspection & Debugging
    docker logs <id>                      # view logs
    docker logs -f <id>                   # follow logs
    docker exec -it <id> bash            # shell into running container
    docker exec -it <id> sh              # for alpine/minimal images
    docker inspect <id>                   # full JSON metadata
    docker stats                          # live resource usage
    docker top <id>                       # processes inside container

## Volumes
    docker volume ls                      # list volumes
    docker volume create mydata           # create named volume
    docker volume rm mydata               # remove volume
    docker volume inspect mydata          # inspect volume

## Networks
    docker network ls                     # list networks
    docker network create mynet           # create network
    docker network inspect mynet          # inspect network
    docker run --network mynet nginx      # attach to network

## Docker Compose
    docker compose up -d                  # start services detached
    docker compose down                   # stop and remove containers
    docker compose logs -f                # follow logs
    docker compose ps                     # list compose services
    docker compose exec web bash          # shell into service

## Cleanup
    docker system prune                   # remove all unused resources
    docker system prune -a                # include unused images
    docker container prune                # remove all stopped containers
