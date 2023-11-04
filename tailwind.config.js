/** @type {import('tailwindcss').Config} */
const colors = require("tailwindcss/colors");

module.exports = {
  content: ["views/**/*.{html,js}"],
  safelist: ["input-error", "input-success"],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
  daisyui: {
    themes: [
      "wireframe",
      "black",
      {
        dark: {
          ...require("daisyui/src/theming/themes")["[data-theme=dark]"],
          primary: "#000000",
          secondary: "#03DAC6",
          accent: "#c084fc",
          neutral: "#1b192a",
          "base-100": "#121212",
          info: "#93c5fd",
          success: "#bef264",
          warning: "#fef08a",
          error: "#cf6679",
        },
        light: {
          ...require("daisyui/src/theming/themes")["[data-theme=light]"],
          primary: "#1c1917",
          secondary: "#9ca3af",
          accent: "#c084fc",
          neutral: "#1b192a",
          "base-100": "#0c0a09",
          info: "#93c5fd",
          success: "#bef264",
          warning: "#fef08a",
          error: "#f87171",
        },
      },
    ],
  },
};
