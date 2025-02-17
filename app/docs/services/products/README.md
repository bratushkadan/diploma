# Products microservice documentation

## SEED(s) use cases

- Add/Get/List/Update/Delete for Product Entity
  - Filtering by *sellerId* is an important requirement for the List handler;
  - Point lookups using *productId*.
- Upload/Delete image for Product (limit is 3MiB)

## Run

### Setup env
```sh
TF_OUTPUT=$(../terraform/tf output -json -no-color)
APP_SA_STATIC_KEY_SECRET_ID="$(echo $TF_OUTPUT | jq -cMr .app_sa.value.static_key_lockbox_secret_id)"
SECRET=$(yc lockbox payload get "${APP_SA_STATIC_KEY_SECRET_ID}")
export AWS_ACCESS_KEY_ID="$(echo $SECRET | yq -M '.entries.[] | select(.key == "access_key_id").text_value')"
export AWS_SECRET_ACCESS_KEY="$(echo $SECRET | yq -M '.entries.[] | select(.key == "secret_access_key").text_value')"
```

