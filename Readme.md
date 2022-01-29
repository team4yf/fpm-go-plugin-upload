# fpm-go-plugin-upload

## Configuration

```yaml
upload:
  dir: "public/uploads/"
  field: "upload"
  base: "/uploads/"
  uploadRouter: "/upload"
  accept:
    - "application/octet-stream"
    - "application/json"
    - "application/zip"
    - "application/x-zip-compressed"
    - "image/png"
    - "image/jpeg"
  limit: 100
```