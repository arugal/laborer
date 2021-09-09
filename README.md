Laborer
===========

[![Build](https://github.com/arugal/laborer/workflows/Test/badge.svg?branch=master)](https://github.com/arugal/laborer/actions?query=workflow%3ATest+branch%3Amaster+event%3Apush)
[![E2E](https://github.com/arugal/laborer/workflows/E2E/badge.svg?branch=master)](https://github.com/arugal/laborer/actions?query=workflow%3AE2E+branch%3Amaster+event%3Apush)
[![Package](https://github.com/arugal/laborer/workflows/Package/badge.svg?branch=master)](https://github.com/arugal/laborer/actions?query=event%3Apush+workflow%3APackage)

`Laborer` 是一个 `Kubernetes` 小工具，代替我们完成一些在开发、测试时的重复动作，比如推送镜像后修改 `tag`，修改配置 `configmap` 后重新部署等.

## 功能

+ 基于 `harbor webhook image push` 事件, 更新对应 `deployment.container` 镜像 `tag`。
+ 基于 `github webhook published` 事件，更新对应 `deployment.container` 镜像 `tag`。
+ `configmap` 被修改时重新部署关联的 `deployment`。
+ 创建 `deployment` 时将镜像 `tag` 修改为 [harbor](https://goharbor.io/) 中最新的 `tag`。

## 部署

1. 克隆代码

   `git clone git@github.com:arugal/laborer.git`

2. 创建 `namespace`

   `kubectl create ns laborer-system`

3. 替换镜像（可选步骤）

   默认镜像为 `ghcr.io/arugal/laborer/manager:latest`，如果需要使用自定义镜像，通过设置环境变量 `export IMG=<image name>` 替换。

4. 安装 [cert-manager](https://cert-manager.io/docs/installation/)

6. 部署 `controller`

   `make deploy`

## 使用配置

1. 启用镜像`tag` 更新和 `configmap` 变化重新部署

    `kubectl label ns <namespace name> laborer.enable=true`

     **注意: 镜像 `push` 后更新对应 `tag` 依赖于镜像仓库的回掉事件，请先根据一下方法设置 webhook**
     
     + Harbor Webhook URL：
     
        `http://laborer-webhook-service.laborer-system/webhook-v1alpha1-harbor-image`
        
       如果 `laborer` 和 `harbor` 部署在同一个 `Kubernetes` 集群内，直接使用上述地址，如果 `harbor` 部署在集群外，请先通过 `NodePort`、`Ingress` 或者 `LoadBalance` 暴露服务
   
     + GitHub Webhook URL:
   
        `http://<ip:port>/webhook-v1alpha1-github-package`
   
        需要先将 laborer-webhook-service 的 80 端口暴露至公网。（局域网环境可使用内网穿透）
   
     + `configmap` 关联规则

       1. 拥有相同名称的 `deployment`，假设 `configmap` 名称为 `test-config` 则关联的 `deployment` 为 `test`
       2. 通过 `annotations/laborer.configmap.associate.deployment` 指定的 `deployment` 集合
    
     + 设置 `configmap` 关联的 `deployment` 集合
        
       `kubectl annotate configmaps <configmap name> -n <namespace name> --overwrite laborer.configmap.associate.deployment="[<deployment array>]"`

2. 启用创建 `deployment` 时修改镜像 `tag`
    
    `kubectl label ns <namespace name> laborere.latest-tag=enabled`
     
     **基于 [Tag.PushTime](https://github.com/arugal/laborer/blob/master/pkg/service/repository/types.go) 排序**

## 兼容性通过版本

+ 1.16.x
+ 1.17.x
+ 1.18.x
+ 1.19.x
+ 1.20.x (未加入 e2e 测试)

## 问题排查

1. 查看 `controller` 运行日志

```bash
   kubectl logs -f -l app=laborer -n laborer-system
```

## License

[Apache 2.0](LICENSE)