module.exports = {
  purge: [
    './templates/**.html',
    '*.go'
  ],
  darkMode: "class", // or 'media' or 'class'
  theme: {
    minWidth: {
      '0': '0',
      '5': '5rem',
      '1/4': '25%',
      '1/2': '50%',
      '3/4': '75%',
      'full': '100%',
    },
    extend: {},
  },
  variants: {
    extend: {},
  },
  plugins: [],
}
