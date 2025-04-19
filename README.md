# solr-inplace-poc


## Run Solr

Run solr with docker-compose.

```bash
docker-compose up --build --watch
```

### Remote docker

connect remote with ssh

```bash
ssh -i ~/.ssh/id_ed25519 -oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null -NL localhost:2377:/var/run/docker.sock user@host
```

```bash
export DOCKER_HOST="tcp://localhost:2377"
```
