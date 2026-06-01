// Package workflow provides a registry of validated workflow definitions for
// applications that host more than one Maestro workflow.
//
// Single-workflow apps should keep using [github.com/justinush/maestro/pkg/maestro]
// (Load, Runtime.NewInstance, Runtime.RestoreInstance).
//
// Multi-workflow apps typically load definitions at startup:
//
//	reg, err := workflow.LoadDir("workflows", validate.Options{})
//
// then start or resume runs by workflow identity:
//
//	in, err := reg.NewInstance(key, maestro.InstanceOptions{...})
//	in, err := reg.RestoreInstance(rec, maestro.InstanceOptions{...})
//
// Register all workflows before serving traffic; treat the registry as read-only
// after startup.
package workflow
