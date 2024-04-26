FROM ubuntu:latest

RUN DEBIAN_FRONTEND=noninteractive apt-get update && apt-get upgrade -y && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY build/ubuntu-reportd /app/ubuntu-reportd
RUN mkdir -p /mnt/data

EXPOSE 8080
ENTRYPOINT ["/app/ubuntu-reportd", "-c", "/app/config.yaml"]
