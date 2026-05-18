package main

import "errors"

var (
	// ErrWrongStep means the client called an endpoint that does not match the current workflow step.
	ErrWrongStep = errors.New("wrong step")

	// ErrApplicantNotFound means no applicant record exists for the given run or applicant id.
	ErrApplicantNotFound = errors.New("applicant: not found")
)
