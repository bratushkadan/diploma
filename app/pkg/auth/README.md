# Public/Private key pairs for asymmetric signing

## Generate keys (new)

```sh
openssl ecparam -genkey -name prime256v1 -noout -out private_key.pem
```

```sh
openssl ec -in private_key.pem -pubout -out public_key.pem
```

## Generate keys (old)

```bash
openssl genrsa -out private.key 4096
```

```bash
openssl rsa -in private.key -pubout -out public.key
```
