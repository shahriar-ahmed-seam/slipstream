import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./app/**/*.{ts,tsx}",
    "./components/**/*.{ts,tsx}",
    "./hooks/**/*.{ts,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        ink: {
          950: "#05060B",
          900: "#0B0E1A",
          800: "#11152A",
          700: "#1B2138",
          600: "#262C49",
          500: "#39406A",
        },
        neon: {
          cyan: "#22D3EE",
          violet: "#A78BFA",
          lime: "#A3E635",
          rose: "#FB7185",
          amber: "#F59E0B",
        },
      },
      fontFamily: {
        sans: ["ui-sans-serif", "system-ui", "Inter", "sans-serif"],
        mono: ["ui-monospace", "SFMono-Regular", "Menlo", "monospace"],
      },
      boxShadow: {
        glow: "0 0 30px rgba(34, 211, 238, 0.25)",
      },
    },
  },
  plugins: [],
};

export default config;