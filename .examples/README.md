# Examples

Runnable example programs showing how `gg` can be applied to concrete
problems. The directory is intentionally named `.examples` so the Go
toolchain skips it when matching `./...`; run an example explicitly with:

```sh
go run ./.examples/access_control
go run ./.examples/inverted_index
go run ./.examples/feature_flags
```

| Example                               | What it demonstrates                                              |
|---------------------------------------|-------------------------------------------------------------------|
| [`access_control`](./access_control/) | Role-based access control — intersection of user sets, membership |
| [`inverted_index`](./inverted_index/) | Keyword search as bitmap union/intersection over documents        |
| [`feature_flags`](./feature_flags/)   | Feature-flag cohorts, A/B audiences and saved state on disk       |

