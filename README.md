# qac

`qac` is a Go library to test end to end command line tools.

A test plan is written in YAML format.

Usage in tests:

```go
import (
  "testing"
  "github.com/enr/qac"
)
func TestExecution(t *testing.T) {
  launcher := qac.NewLauncher()
  report := launcher.ExecuteFile(`/path/to/qac.yaml`)
  // Not needed but useful to see what's happening
  reporter := qac.NewTestLogsReporter(t)
  reporter.Publish(report)
  // Fail test if any error is found
  for _ei_, err := range report.AllErrors() {
    t.Errorf(`error %v`, err)
  }
}
```

## License

Apache 2.0 - see LICENSE file.

Copyright 2020 qac contributors
