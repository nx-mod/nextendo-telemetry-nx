# nextendo-telemetry-nx — example usage

`./run.sh` starts the sink (:8472). The console uploads play/error reports
synchronously; if they hang, some applets stall. This answers 200 `{}` to any
request so the upload completes, while nothing is forwarded to Nintendo.

```sh
./run.sh &
./test-sink.sh          # POSTs a fake report -> 200 {}
```

Set `TELEMETRY_DUMP=1` before `./run.sh` to log the bodies it receives.
