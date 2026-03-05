# Story 10.4: CRD Cleanup & Drift Alert

**Status:** done

## Summary
Implemented CRDCleaner for periodic Tekton run cleanup (mock K8s) and DriftDetector for ArgoCD drift detection.

## Deliverables
- `internal/crdclean/cleaner.go` - CRDCleaner, CleanExpiredRuns (TODO mock), StartScheduler
- `internal/crdclean/drift.go` - DriftDetector, CheckDrift, DriftReport
- `internal/crdclean/cleaner_test.go` - tests
- `internal/crdclean/drift_test.go` - tests

## Notes
- K8s deletion is mock with TODO for real Tekton PipelineRun cleanup
- DriftDetector uses ArgoClient interface (sync/health); wire real argocd client via adapter
