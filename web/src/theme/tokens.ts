// Design tokens aligned with Stitch project "zcid Dashboard Overview"
// (projects/728687561751426614). Source of truth: designMd + rendered screens.
// Creative north star: "Living Blueprint" — tonal layering, no-line rule,
// Apple-editorial authority.

export const zcidTheme = {
  // Brand
  primary: '#0057c2',
  primaryContainer: '#006ef2',
  primaryFixed: '#d9e2ff',
  primaryFixedDim: '#afc6ff',
  onPrimary: '#ffffff',
  onPrimaryFixedVariant: '#004398',

  // Override (Arco Design default alias)
  primaryOverride: '#1677FF',

  // Surface hierarchy (light main)
  surface: '#f8f9fb',
  surfaceContainerLowest: '#ffffff',
  surfaceContainerLow: '#f2f4f6',
  surfaceContainer: '#edeef0',
  surfaceContainerHigh: '#e7e8ea',
  surfaceContainerHighest: '#e1e2e4',
  surfaceDim: '#d9dadc',
  surfaceBright: '#f8f9fb',

  // Sidebar (dark rail per rendered screens)
  sidebarBg: '#0f1418',
  sidebarBgAlt: '#171c20',
  sidebarSurface: '#1b2024',
  sidebarSurfaceHigh: '#262b2f',
  sidebarFg: '#dfe3e9',
  sidebarMuted: '#8a919f',
  sidebarOutline: '#404753',
  sidebarActiveBg: '#2492ff',

  // On-surface text
  onSurface: '#191c1e',
  onSurfaceVariant: '#414755',
  onBackground: '#191c1e',

  // Outline / ghost borders (use at 15-20% opacity only)
  outline: '#727786',
  outlineVariant: '#c1c6d7',

  // Status
  success: '#0057c2',
  successText: '#004398',
  successContainer: '#d9e2ff',
  error: '#ba1a1a',
  errorContainer: '#ffdad6',
  onErrorContainer: '#93000a',
  warning: '#9e3d00',
  warningContainer: '#ffdbcc',
  onWarningContainer: '#351000',
  runningGlow: 'rgba(0, 87, 194, 0.3)',

  // Log viewer (dark terminal)
  logBg: '#1E1E1E',
  logBgAlt: '#262b2f',
  logFg: '#D4D4D4',
  logLineNum: '#858585',
  logInfo: '#74C0FC',
  logSuccess: '#69DB7C',
  logWarn: '#FFC107',
  logError: '#FF6B6B',

  // Radii (8px logic)
  radiusSm: '6px',
  radiusMd: '8px',
  radiusLg: '12px',
  radiusXl: '16px',
  radiusCapsule: '24px',
  radiusFull: '9999px',

  // Spacing (8px grid)
  space1: '4px',
  space2: '8px',
  space3: '12px',
  space4: '16px',
  space5: '20px',
  space6: '24px',
  space8: '32px',
  space10: '40px',
  space12: '48px',
  space16: '64px',

  // Typography
  fontDisplay: "'Manrope', system-ui, -apple-system, sans-serif",
  fontBody: "'Inter', system-ui, -apple-system, sans-serif",
  fontMono: "'JetBrains Mono', 'Fira Code', Consolas, monospace",

  // Signature gradient (primary → primary_container, 135deg)
  gradientPrimary: 'linear-gradient(135deg, #0057c2 0%, #006ef2 100%)',

  // Ambient shadows (blue-tinted, not grey)
  shadowSoft: '0 1px 2px rgba(0, 87, 194, 0.04)',
  shadowFloat: '0 8px 24px rgba(0, 87, 194, 0.06)',
  shadowModal: '0 20px 40px rgba(0, 87, 194, 0.08)',

  // Glass
  glassFill: 'rgba(255, 255, 255, 0.72)',
  glassBlur: '20px',
} as const;
