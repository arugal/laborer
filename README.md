Laborer
===========

# Features

# Quick start

```shell script
# generate cert
bash hack/webhook-create-signed-cert.sh

# replace caBundle
cat config/default/webhook_ca_bundle_patch_template.yaml | bash hack/webhook-patch-ca-bundle.sh > config/default/webhook_ca_bundle_patch.yaml

# apply
cd config/default
kustomize build | kubectl apply -f -
```