# prmtvs

Some primitives structs.

## SKM - sorted keys map

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
