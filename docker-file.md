# dockerfile使用说明

## my-redis
```Shell
docker rm -f my-redis
```
```Shell
docker build -t my-redis -f redis.dockerfile .
```
```Shell
docker run -d \
  --name my-redis \
  -p 6379:6379 \
  my-redis

```
```Shell
docker logs -f --tail 100 my-redis
```


## armylong-go

```Shell
docker rm -f armylong-go
```
```Shell
docker build -t armylong-go -f go.dockerfile .
```
```Shell
docker run -d \
  --name armylong-go \
  -p 8080:8080 \
  -e REDIS_HOST=host.docker.internal \
  -e REDIS_PORT=6379 \
  -e REDIS_PASSWORD=123456 \
  -e FEISHU_ARMYLONG_APP_ID=cli_a94dc0fc84f6dbdd___zzl \
  -e FEISHU_ARMYLONG_APP_SECRET=HByUfMitn7ThWxAydsJP6oz2RikAmSXC___zzl \
  armylong-go serve
```
```Shell
docker logs -f --tail 100 armylong-go
```
