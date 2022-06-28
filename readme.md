# demo_tracing

## start:
```shell
docker-compose up --build
```

### demo:

- jaeger will be available at http://127.0.0.1:16686/
- application A will be available at http://127.0.0.1:8091/
- application B will be available at http://127.0.0.1:8092/

To see the result, you need to send a GET request to http://127.0.0.1:8091/hello.

After that, you can see the trace at http://127.0.0.1:16686/search
