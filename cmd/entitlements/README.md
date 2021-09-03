## Introduction
This CLI is used to generate new entitlements intended to be used by WeGO EE customers. The entitlement is generated in the form of a Kubernetes secret which packages a signed (but not encrypted) JWT token. The token itself specifies the expiry date of the entitlement.

The tool is currently using an Ed25519 key pair of which the public key is bundled with WeGO EE and is used to verify the validity and integrity of the entitlements used by customers. The private key is kept in 1pass.

### How to use
Run the following command to generate an entitlement for mail@customer.org for 1 year using the private key `pair.pem`:
```
entitlements generate -p ./pair.pem -c mail@customer.org -y 1 > secret.yaml
```

###  How to inspect an entitlement
Read the entitlement and base64 decode it:
```
kubectl get secrets/wego-ee-entitlement --template={{.data.entitlement}} | base64 -D
```
Then copy it into https://jwt.io/#debugger-io and inspect its payload.


### How to create a new Ed25519 key pair
Use `openssl` and specify the ED25519 algorithm option:
```bash
openssl genpkey -algorithm ED25519 > pair.pem
```
The generated PEM-encoded file should be stored in 1-pass.
Extract the public key from the newly generated key pair:
```
openssl pkey -in pair.pem -pubout > public.pem
```
The generated PEM-encoded file should be copied into the WeGO EE source code.

Note that this process should ideally happen only once for WeGO EE as it affects the bundled public key that is shipped with the product.