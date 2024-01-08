# prmtvs

Some primitives structures:
- `plx` - channels plexus
- `skm` - sorted keys map

## Plx - channels plexus

Struct is a synchronization primitive based on the queue of channels. It helps to link several inputs to several 
 outputs. 

## Skm - sorted keys map

Struct is based on hash map and sorted slice of all keys. It allows get values by the string or by the index.
 It has two implementation: thread-safe and not 

# contribution

Do not forget to format the source code using hot `make` command: 

```bash
make fmt
```

Don't forget to run test before any PRs:

```bash
make test
```
