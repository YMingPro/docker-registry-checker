
# Docker Registry Checker

> 一个检测Docker镜像源的工具
> 
> 默认情况下取工作目录下的docker.txt文件进行检查，如果文件不存在则会从当前仓库拉取最新的数据

### 命令说明：
- `-l` 参数来筛选只显示成功的结果
- `-timeout` 指定请求超时时间（秒）
- `-update` 强制从GitHub更新docker.txt
- `-workers` 并发worker的数量

### 修改镜像源步骤
```shell
# 使用vim修改daemon.json 文件中的registry-mirrors字段
vim /etc/docker/daemon.json

# 或者使用 cat 命令将以下内容写入 /etc/docker/daemon.json 文件
cat >/etc/docker/daemon.json <<EOF
{
  "registry-mirrors": [
    "https://registry-1.docker.io",
  ]
}
EOF

systemctl daemon-reload && systemctl restart docker
```

### 版本说明：
```
todo
```

感谢Goland提供的支持

感谢dockerproxy.xiaoxuan6.us.kg和http://status.kggzs.cn/status/docker提供的数据

