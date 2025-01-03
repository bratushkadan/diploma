# Public/Private key pairs for asymmetric signing

## Generate keys

```bash
openssl genrsa -out private.key 2048
```

```bash
openssl rsa -in private.key -pubout -out public.key
```
