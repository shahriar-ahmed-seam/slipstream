export const palette = {
  bg: "#05060B",
  surface: "#0B0E1A",
  surfaceAlt: "#11152A",
  border: "#1B2138",
  text: "#E5E7EB",
  textMuted: "#94A3B8",
  accent: "#22D3EE",
  accent2: "#A78BFA",
  ok: "#A3E635",
  warn: "#F59E0B",
  err: "#FB7185",
} as const;

export type Palette = typeof palette;