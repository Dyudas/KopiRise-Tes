/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        espresso: '#261B15',
        espressoDark: '#17100D',
        caramel: '#C4844E',
        caramelHover: '#A76D3B',
        creamWarm: '#FAF6F0',
        creamCard: '#FFFFFF'
      },
      fontFamily: {
        heading: ['"Plus Jakarta Sans"', 'sans-serif'],
        sans: ['Inter', 'sans-serif']
      }
    }
  },
  plugins: []
};
