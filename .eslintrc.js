module.exports = {
  root: true,
  extends: [
    "eslint:recommended",
    "plugin:@typescript-eslint/recommended",
    "plugin:import/errors",
    "plugin:import/typescript",
    "plugin:import/warnings",
    "plugin:jest/recommended",
    "plugin:jsx-a11y/recommended",
    "plugin:react-hooks/recommended",
    "plugin:react/recommended",
    "plugin:testing-library/react",
  ],
  rules: {
    "import/default": [0],
    "import/no-named-as-default-member": [0],
    "import/no-named-as-default": [0],
    "import/no-unresolved": [0],
    "import/order": [
      2,
      {
        alphabetize: {
          order: "asc",
          caseInsensitive: true,
        },
        groups: [
          "builtin",
          "external",
          "internal",
          "parent",
          "sibling",
          "index",
          "object",
          "type",
        ],
      },
    ],

    "@typescript-eslint/explicit-module-boundary-types": [0],
    "@typescript-eslint/no-unused-vars": [0],
    "@typescript-eslint/ban-types": [
      "error",
      {
        types: {
          Object: false,
          "{}": false,
        },
        extendDefaults: true,
      },
    ],
    "@typescript-eslint/no-explicit-any": [0],
    "@typescript-eslint/ban-ts-comment": [0],
    "@typescript-eslint/switch-exhaustiveness-check": [2],
    "@typescript-eslint/no-non-null-assertion": [0],

    // testing stuff
    "testing-library/no-node-access": [0],
    "testing-library/no-unnecessary-act": [0],
    "testing-library/render-result-naming-convention": [0],
    "testing-library/prefer-screen-queries": [0],
    "jest/expect-expect": [0],

    // Let you omit `import React` from files
    "react/react-in-jsx-scope": [0],
    "react/jsx-uses-react": [0],
    // Let you use quotes etc in JSX blocks
    "react/no-unescaped-entities": [0],
    // FIXME: lets switch this on for extra safety.
    // We get a runtime check on this too but it's nice to have it in the linter
    "react/jsx-key": [0],

    // FIXME: nice to address these
    "jsx-a11y/label-has-associated-control": [0],
    "jsx-a11y/click-events-have-key-events": [0],
    "jsx-a11y/no-static-element-interactions": [0],
    "jsx-a11y/no-autofocus": [0],
  },
  settings: {
    react: {
      version: "detect",
    },
  },
  ignorePatterns: [
    "rpc",
    "api",
    "assets",
    "cluster-services",
    "fonts",
    "node_modules",
    ".d.ts",
    ".pb.ts",
    "setupProxy.js",
    "test-utils.tsx",
    "reportWebVitals.ts",
  ],
  parserOptions: {
    project: "./tsconfig.json",
  },
};
