export const colors = {
  // Primary - Red tones (replacing Anthropic's orange/terracotta)
  primary: '#DC2626',
  primaryLight: '#FEE2E2',
  primaryDark: '#991B1B',
  primaryMuted: '#FECACA',

  // Neutrals - Anthropic style (warm grays)
  background: '#FAFAF9',
  surface: '#FFFFFF',
  surfaceElevated: '#F5F5F4',

  // Text
  textPrimary: '#1C1917',
  textSecondary: '#57534E',
  textTertiary: '#A8A29E',
  textOnPrimary: '#FFFFFF',

  // Borders
  border: '#E7E5E4',
  borderFocused: '#DC2626',

  // Status
  success: '#16A34A',
  warning: '#D97706',
  error: '#DC2626',
  info: '#2563EB',

  // Map specific
  busMarker: '#DC2626',
  routePath: '#DC262680',
  campusBounds: '#DC262620',
  campusBorder: '#DC2626',

  // Overlays
  overlay: 'rgba(0, 0, 0, 0.5)',
  shimmer: '#E7E5E4',
} as const;
