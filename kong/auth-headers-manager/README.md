# auth-headers-manager

A Kong plugin that will add Authorization header from a cookie.

## How it works

When enabled, this plugin will populate authorization headers `Authorization` and `X-Okapi-Token` from
cookie `folioAccessToken` according to the configuration

## Configuration

```bash
curl -X POST http://localhost:8001/apis/{api_id}/plugins \
--data "name=auth-headers-manager" \
--data "config.set_okapi_header=false" \
--data "config.set_authorization_header=true"
```

### Configuration parameters

| Form Parameter             | Required   | Default Value | Description                                                                          |
|----------------------------|------------|:--------------|--------------------------------------------------------------------------------------|
| `set_okapi_header`         | *optional* | true          | Defines if X-Okapi-Token must be populated if folioAccessToken is present in cookies |
| `set_authorization_header` | *optional* | false         | Defines if X-Okapi-Token must be populated if folioAccessToken is present in cookies |
