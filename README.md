evm

`zig version`

Build and run the optimized native binary:

```sh
bash install.sh
./run.sh --count 1000 --workers 2
```

Workers are CPU threads. The default uses all CPUs visible to the process; set
`--workers` explicitly when running alongside other workloads. `--worker` is
kept as a compatibility alias.
