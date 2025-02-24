
# Docker Registry Checker

> 一个检测Docker镜像源的脚本工具 (Windows / Linux)
> 
> 默认情况下取工作目录下的docker.txt文件进行检查，如果文件不存在则会从当前仓库拉取最新的数据

### 功能

- ✅支持批量检测registry是否可用
- ✅支持Linux检测后批量替换或指定替换registry
- ⬜Linux一键安装docker和docker-compose[计划中]

### 使用
#### Linux下的一键执行命令:

```bash
curl -L https://github.com/YMingPro/docker-registry-checker/releases/latest/download/docker-registry-checker -o docker-registry-checker && chmod +x docker-registry-checker && ./docker-registry-checker
```
或者
```bash
wget https://github.com/YMingPro/docker-registry-checker/releases/latest/download/docker-registry-checker && chmod +x docker-registry-checker && ./docker-registry-checker
```

#### Windows下的一键执行命令:
```cmd
curl -L https://github.com/YMingPro/docker-registry-checker/releases/latest/download/docker-registry-checker.exe -o docker-registry-checker.exe && docker-registry-checker.exe
```

### 可选参数说明：
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
