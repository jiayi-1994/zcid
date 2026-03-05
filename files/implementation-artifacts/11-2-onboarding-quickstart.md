# Story 11.2: First-time Onboarding & Quick Start

**Status:** done

## Summary
Implemented OnboardingCard component for first-time users with "Your projects", "Create your first pipeline", and "Documentation" links, plus "Don't show again" that persists via localStorage.

## Deliverables
- `web/src/components/onboarding/OnboardingCard.tsx` - OnboardingCard component with localStorage key `zcid_onboarding_dismissed`
- `web/src/pages/dashboard/DashboardPage.tsx` - Integrated OnboardingCard above dashboard when not dismissed

## Notes
- Shows when `!localStorage.getItem('zcid_onboarding_dismissed')`
- "Documentation" links to GitHub repo; can be updated to docs URL when available
- onDismiss callback updates parent state to hide card immediately
