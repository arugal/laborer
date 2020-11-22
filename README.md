Laborer
===========

[![Build](https://github.com/arugal/laborer/workflows/Test/badge.svg?branch=master)](https://github.com/arugal/laborer/actions?query=workflow%3ATest+branch%3Amaster+event%3Apush)
[![E2E](https://github.com/arugal/laborer/workflows/E2E/badge.svg?branch=master)](https://github.com/arugal/laborer/actions?query=workflow%3AE2E+branch%3Amaster+event%3Apush)

`Laborer` 是一个 `Kubernetes` 小工具，代替我们完成一些在开发、测试时的重复动作，比如更新 `tag`，重新部署。

## 功能

+ 基于 `harbor webhook push` 事件, 更新 `deployment` 镜像 `tag`
+ `configmap` 变化时重新部署关联的 `deployment`
+ 创建 `deployment` 时将镜像 `tag` 修改为 `harbor` 中最新的 `tag`


## 使用配置

1. 启用镜像`tag` 更新和 `configmap` 变化重新部署

    `kubectl label ns <namespace name> laborer.enable=true`

     **注意: tag更新依赖 `harbor webhook` 的回调事件，除了配置 `label` 还需要在对应的仓库配置 `webhook` **
     
2. `configmap` 关联规则

    + 拥有相同名称的 `deployment`。假设 `configmap` 名称为 `test-config` 则关联的 `deployment` 为 `test`
    + 通过 `annotations/laborer.configmap.associate.deployment` 指定的 `deployment` 集合

3. 指定 `configmap` 关联的 `deployment` 集合

    `kubectl annotate configmaps <configmap name> -n <namespace name> --overwrite laborer.configmap.associate.deployment="[<deployment array>]"`

4. 启用创建 `deployment` 时修改镜像 `tag`

    `kubectl label ns <namespace name> laborere.latest-tag=enabled`

## Kubernetes 适配版本

+ 1.16.x
+ 1.17.x
+ 1.18.x
+ 1.19.x

## 部署

1. 克隆代码

    `git clone git@github.com:arugal/laborer.git`

2. 创建 `namespace`
    
    `kubectl create ns laborer-system`
    
3. 替换镜像（可选步骤） 

    默认镜像为 `docker.pkg.github.com/arugal/laborer/manager:latest`，如果需要使用自定义镜像，通过设置环境变量 `export IMG=<image name>` 替换。

4. 创建 `webhook` 并 `approve` 证书

    `bash hack/webhook-create-signed-cert.sh --namespace laborer-system --service laborer-webhook-service --secret laborer-webhook-server-cert`

5. 替换 `caBundle` 属性

    `cat config/default/webhook_ca_bundle_patch_template.yaml | bash hack/webhook-patch-ca-bundle.sh > config/default/webhook_ca_bundle_patch.yaml`

6. 部署 `controller`

    `make deploy`

## 问题排查

1. 查看 `controller` 运行日志

```bash
    kubectl logs -f $(kubectl get pods -n laborer-system -o jsonpath="{.items[0].metadata.name}") -n laborer-system 
```

## License

[Apache 2.0](LICENSE)