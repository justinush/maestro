package main

import "fmt"

// FakeVendorCheckLiveness simulates an external liveness vendor (Onfido/Sumsub-style).
// The workflow still records the outcome via the stub action on run-liveness-check;
// this hook shows where your app would call the vendor before Maestro advances.
func FakeVendorCheckLiveness(applicantID string) error {
	fmt.Printf("vendor: liveness check queued for applicant %s (offline stub)\n", applicantID)
	return nil
}
