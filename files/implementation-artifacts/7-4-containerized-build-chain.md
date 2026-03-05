# Story 7.4: Containerized Build Chain

**Status:** done

## Summary
- `pkg/tekton/buildchain.go`: BuildChainGenerator with GenerateContainerBuildSteps
- `pkg/tekton/buildchain_test.go`: Tests for container build steps
- `pkg/tekton/traditional.go`: GenerateTraditionalBuildSteps
- `pkg/tekton/traditional_test.go`: Tests for traditional build steps

## Steps
1. git-clone - Clone repo using alpine/git
2. build - Compile using specified build image
3. kaniko-build-push - Build+push Docker image using gcr.io/kaniko-project/executor

## Retry
- RetryAnnotationKey/Value constants for max 3 retries (applied when wrapping in PipelineTask)
