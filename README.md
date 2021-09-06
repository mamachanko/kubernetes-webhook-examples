# Kubernetes Webhook Examples
> Go & Java ftw!

Deploy a Java and a Golang application which serve validating webhooks for `Thing` custom resources.

Get started:
```sh
# create certificates et al.
go run certs.go 
# create a cluster
kind create cluster --config kind-config.yaml
# continuously deploy and watch logs
skaffold dev --cleanup=false
```

Test:
```sh
# create a `Thing` and see both applications validating it
cat <<EOF | kubectl apply -f -
apiVersion: mamachanko.com/v1alpha1
kind: Thing
metadata:
  name: test
EOF
```
