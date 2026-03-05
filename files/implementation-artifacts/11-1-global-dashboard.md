# Story 11.1: Global Dashboard

**Status:** done

## Summary
Implemented global dashboard with project cards, stats section (total projects, pipelines, recent runs success/fail), and environment health summary per project.

## Deliverables
- `web/src/services/dashboard.ts` - fetchDashboardData(), DashboardProject, DashboardStats types
- `web/src/pages/dashboard/DashboardPage.tsx` - Project cards (name, last run status, env health), stats grid, card click navigates to project detail
- `web/src/App.tsx` - Dashboard route already uses DashboardPage (no change needed)

## Notes
- Dashboard aggregates data from projects, pipelines, runs, deployments APIs
- Status badges: healthy=green, degraded=orange, failed=red
- Last run status from first active pipeline per project
