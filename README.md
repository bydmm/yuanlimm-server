# 援力满满 - 二次元虚拟股市 - 高速验证服务

## 用途

援力满满高速验证服务，开源接受检阅。不过由于援力满满的REDIS不对外开放，所以本工具没有实际用途。

## 快速启动
```bash
docker run daocloud.io/bydmm/yuanlimm-server -p 8000:8000 \
    --env REDIS_ADDR="redis:6379"
    --env REDIS_PW="password"
    --env REDIS_DB="1"
    --env WISH_URL="http://yuanlimm-server.yuanlimm.com/api/super_wishs"
```

