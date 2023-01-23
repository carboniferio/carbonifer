# Testing Carbonifer

This folder contains test data. 
Tests are defined as `/internal/**/*_test.go` files.

To run them properly you need to:

```bash
export "GOOGLE_APPLICATION_CREDENTIALS": "<path to gcp token json file>"
go tests ./...
```

You can also skip tests requiring credentials:

```bash
SKIP_WITH_CREDENTIALS=true go test -v ./...
```