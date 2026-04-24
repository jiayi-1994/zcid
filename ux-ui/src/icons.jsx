// Lucide-inspired stroke icons. All 16px default, 1.75 stroke.
const Icon = ({ d, size = 14, stroke = 1.75, fill = 'none', ...rest }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24"
    fill={fill} stroke="currentColor" strokeWidth={stroke} strokeLinecap="round" strokeLinejoin="round" {...rest}>
    {typeof d === 'string' ? <path d={d} /> : d}
  </svg>
);

const I = {
  // nav
  grid: (p) => <Icon {...p} d="M3 3h7v7H3zM14 3h7v7h-7zM3 14h7v7H3zM14 14h7v7h-7z" />,
  folder: (p) => <Icon {...p} d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />,
  users: (p) => <Icon {...p} d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2M9 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" />,
  key: (p) => <Icon {...p} d="M21 2l-9.6 9.6M15.5 7.5l3 3M10.34 14.66a5 5 0 1 1-7.07-7.07 5 5 0 0 1 7.07 7.07z" />,
  plug: (p) => <Icon {...p} d="M12 22v-5M9 8V2M15 8V2M7 8h10v4a5 5 0 0 1-10 0z" />,
  shield: (p) => <Icon {...p} d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />,
  settings: (p) => <Icon {...p} d="M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6zM19.4 15a1.7 1.7 0 0 0 .3 1.8l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-1.8-.3 1.7 1.7 0 0 0-1 1.5V21a2 2 0 0 1-4 0v-.1a1.7 1.7 0 0 0-1.1-1.5 1.7 1.7 0 0 0-1.8.3l-.1.1A2 2 0 1 1 4.3 17l.1-.1a1.7 1.7 0 0 0 .3-1.8 1.7 1.7 0 0 0-1.5-1H3a2 2 0 0 1 0-4h.1a1.7 1.7 0 0 0 1.5-1.1 1.7 1.7 0 0 0-.3-1.8l-.1-.1A2 2 0 1 1 7 4.3l.1.1a1.7 1.7 0 0 0 1.8.3h0a1.7 1.7 0 0 0 1-1.5V3a2 2 0 0 1 4 0v.1a1.7 1.7 0 0 0 1 1.5 1.7 1.7 0 0 0 1.8-.3l.1-.1A2 2 0 1 1 19.7 7l-.1.1a1.7 1.7 0 0 0-.3 1.8v0a1.7 1.7 0 0 0 1.5 1H21a2 2 0 0 1 0 4h-.1a1.7 1.7 0 0 0-1.5 1z" />,
  // actions
  plus: (p) => <Icon {...p} d="M12 5v14M5 12h14" />,
  minus: (p) => <Icon {...p} d="M5 12h14" />,
  target: (p) => <Icon {...p} d="M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18zM12 17a5 5 0 1 0 0-10 5 5 0 0 0 0 10zM12 13a1 1 0 1 0 0-2 1 1 0 0 0 0 2z" />,
  search: (p) => <Icon {...p} d="M11 19a8 8 0 1 0 0-16 8 8 0 0 0 0 16zM21 21l-4.35-4.35" />,
  refresh: (p) => <Icon {...p} d="M3 12a9 9 0 0 1 15-6.7L21 8M21 3v5h-5M21 12a9 9 0 0 1-15 6.7L3 16M3 21v-5h5" />,
  play: (p) => <Icon {...p} d="M6 4l15 8-15 8z" fill="currentColor" stroke="currentColor" />,
  pause: (p) => <Icon {...p} d="M6 4h4v16H6zM14 4h4v16h-4z" fill="currentColor" stroke="currentColor" />,
  edit: (p) => <Icon {...p} d="M12 20h9M16.5 3.5a2.12 2.12 0 1 1 3 3L7 19l-4 1 1-4z" />,
  trash: (p) => <Icon {...p} d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2M10 11v6M14 11v6" />,
  copy: (p) => <Icon {...p} d="M9 9h10a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H9a2 2 0 0 1-2-2V11a2 2 0 0 1 2-2zM5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />,
  drag: (p) => <Icon {...p} d="M9 6h.01M9 12h.01M9 18h.01M15 6h.01M15 12h.01M15 18h.01" fill="currentColor" stroke="currentColor" strokeWidth={3} />,
  more: (p) => <Icon {...p} d="M12 13a1 1 0 1 0 0-2 1 1 0 0 0 0 2zM19 13a1 1 0 1 0 0-2 1 1 0 0 0 0 2zM5 13a1 1 0 1 0 0-2 1 1 0 0 0 0 2z" fill="currentColor" />,
  check: (p) => <Icon {...p} d="M20 6L9 17l-5-5" />,
  x: (p) => <Icon {...p} d="M18 6L6 18M6 6l12 12" />,
  // arrows
  arrR: (p) => <Icon {...p} d="M5 12h14M13 5l7 7-7 7" />,
  arrL: (p) => <Icon {...p} d="M19 12H5M11 5l-7 7 7 7" />,
  arrD: (p) => <Icon {...p} d="M12 5v14M5 13l7 7 7-7" />,
  chevR: (p) => <Icon {...p} d="M9 6l6 6-6 6" />,
  chevD: (p) => <Icon {...p} d="M6 9l6 6 6-6" />,
  chevL: (p) => <Icon {...p} d="M15 6l-6 6 6 6" />,
  // pipeline
  zap: (p) => <Icon {...p} d="M13 2L3 14h8l-1 8 11-12h-8z" />,
  branch: (p) => <Icon {...p} d="M6 3v12M18 9a3 3 0 1 0 0 6 3 3 0 0 0 0-6zM6 21a3 3 0 1 0 0-6 3 3 0 0 0 0 6zM6 15a9 9 0 0 0 9-6" />,
  gitCommit: (p) => <Icon {...p} d="M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6zM3 12h6M15 12h6" />,
  layers: (p) => <Icon {...p} d="M12 2L2 7l10 5 10-5zM2 17l10 5 10-5M2 12l10 5 10-5" />,
  server: (p) => <Icon {...p} d="M3 5h18v6H3zM3 13h18v6H3zM7 8h0M7 16h0" />,
  cube: (p) => <Icon {...p} d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16zM3.27 6.96L12 12l8.73-5.04M12 22V12" />,
  rocket: (p) => <Icon {...p} d="M4.5 16.5c-1.5 1.26-2 5-2 5s3.74-.5 5-2c.71-.84.7-2.13-.09-2.91a2.18 2.18 0 0 0-2.91-.09zM12 15l-3-3a22 22 0 0 1 2-3.95A12.88 12.88 0 0 1 22 2c0 2.72-.78 7.5-6 11a22 22 0 0 1-4 2zM9 12H4s.55-3.03 2-4c1.62-1.08 5 0 5 0M12 15v5s3.03-.55 4-2c1.08-1.62 0-5 0-5" />,
  bell: (p) => <Icon {...p} d="M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9M10 21a2 2 0 0 0 4 0" />,
  terminal: (p) => <Icon {...p} d="M4 17l6-6-6-6M12 19h8" />,
  code: (p) => <Icon {...p} d="M16 18l6-6-6-6M8 6l-6 6 6 6" />,
  book: (p) => <Icon {...p} d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20V2H6.5A2.5 2.5 0 0 0 4 4.5v15zM4 19.5A2.5 2.5 0 0 0 6.5 22H20v-5" />,
  logOut: (p) => <Icon {...p} d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9" />,
  home: (p) => <Icon {...p} d="M3 12L12 3l9 9M5 10v10h14V10" />,
  calendar: (p) => <Icon {...p} d="M3 6a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2zM16 2v4M8 2v4M3 10h18" />,
  download: (p) => <Icon {...p} d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3" />,
  external: (p) => <Icon {...p} d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6M15 3h6v6M10 14L21 3" />,
  link: (p) => <Icon {...p} d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.72-1.71" />,
  filter: (p) => <Icon {...p} d="M22 3H2l8 9.46V19l4 2v-8.54z" />,
  clock: (p) => <Icon {...p} d="M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20zM12 6v6l4 2" />,
  alert: (p) => <Icon {...p} d="M10.3 3.86L1.8 18a2 2 0 0 0 1.7 3h17a2 2 0 0 0 1.7-3L13.7 3.86a2 2 0 0 0-3.4 0zM12 9v4M12 17h.01" />,
  user: (p) => <Icon {...p} d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2M12 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8z" />,
  lock: (p) => <Icon {...p} d="M5 11h14v10H5zM7 11V7a5 5 0 0 1 10 0v4" />,
  eye: (p) => <Icon {...p} d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8zM12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z" />,
  cloud: (p) => <Icon {...p} d="M17.5 19a4.5 4.5 0 1 0-1.4-8.78A7 7 0 1 0 4 15" />,
  sync: (p) => <Icon {...p} d="M21 2v6h-6M3 12a9 9 0 0 1 15-6.7L21 8M3 22v-6h6M21 12a9 9 0 0 1-15 6.7L3 16" />,
  globe: (p) => <Icon {...p} d="M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20zM2 12h20M12 2a15 15 0 0 1 4 10 15 15 0 0 1-4 10 15 15 0 0 1-4-10 15 15 0 0 1 4-10z" />,
  heart: (p) => <Icon {...p} d="M20.8 4.6a5.5 5.5 0 0 0-7.8 0L12 5.7l-1-1.1a5.5 5.5 0 1 0-7.8 7.8l1.1 1.1L12 21l7.8-7.8 1-1.1a5.5 5.5 0 0 0 0-7.8z" />,
  list: (p) => <Icon {...p} d="M8 6h13M8 12h13M8 18h13M3 6h.01M3 12h.01M3 18h.01" />,
  file: (p) => <Icon {...p} d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8zM14 2v6h6" />,
};

Object.assign(window, { Icon, I });
