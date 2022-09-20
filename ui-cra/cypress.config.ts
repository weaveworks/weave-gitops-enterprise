import { defineConfig } from 'cypress';

export default defineConfig({
  e2e: {
    setupNodeEvents(on, config) {
      // implement node event listeners here
    },
  },

  component: {
    viewportWidth: 900,
    viewportHeight: 500,
    devServer: {
      framework: 'create-react-app',
      bundler: 'webpack',
    },
  },
});
