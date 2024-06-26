# kong

## Introduction

A docker image for kong.

## Environment Variables

| Name                                         |       Default value        | Suggested value | Required | Description                                                                                                                          |
|:---------------------------------------------|:--------------------------:|:---------------:|:--------:|:-------------------------------------------------------------------------------------------------------------------------------------|
| KONG_NGINX_HTTPS_LARGE_CLIENT_HEADER_BUFFERS |             -              |     4 200k      |   true   | Set buffer size for large headers to embedded nginx. (https)                                                                         |
| KONG_NGINX_HTTP_LARGE_CLIENT_HEADER_BUFFERS  |             -              |     4 200k      |   true   | Set buffer size for large headers to embedded nginx. (http)                                                                          |
