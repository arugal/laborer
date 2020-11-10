Laborer
===========

# Features

# Quick start

```shell script
# generate cert
bash hack/webhook-create-signed-cert.sh --namespace laborer-system --service laborer-webhook-service --secret laborer-webhook-server-cert

# replace caBundle
cat config/default/webhook_ca_bundle_patch_template.yaml | bash hack/webhook-patch-ca-bundle.sh > config/default/webhook_ca_bundle_patch.yaml

# apply
make deploy
```