# Story 11.3: Global List Filters & Search

**Status:** done

## Summary
Implemented reusable filter infrastructure (useQueryFilters hook, ListFilters component) and added search/filter to PipelineListPage and PipelineRunListPage with URL param sync.

## Deliverables
- `web/src/hooks/useQueryFilters.ts` - useQueryFilters<T>(defaults): [T, (updates: Partial<T>) => void]
- `web/src/components/common/ListFilters.tsx` - Reusable filter bar (search input + select filters)
- `web/src/pages/projects/pipelines/PipelineListPage.tsx` - Search by pipeline name, status filter, URL sync
- `web/src/pages/projects/pipelines/PipelineRunListPage.tsx` - Filters: status, trigger type; Search: commit SHA; URL sync
- `web/src/services/pipelineRun.ts` - PipelineRunSummary extended with gitCommit? for future backend support

## Notes
- Filters sync to URL query params via useSearchParams
- Commit SHA search filters when backend includes gitCommit in list response (extended type for future)
- Client-side filtering within fetched page data
