import js from '@eslint/js'
import prettierConfig from 'eslint-config-prettier'
import jsxA11yPlugin from 'eslint-plugin-jsx-a11y'
import onlyWarn from 'eslint-plugin-only-warn'
import reactPlugin from 'eslint-plugin-react'
import reactHooksPlugin from 'eslint-plugin-react-hooks'
import simpleImportSortPlugin from 'eslint-plugin-simple-import-sort'
import eslintPluginAstro from 'eslint-plugin-astro'
import globals from 'globals'
import tseslint from 'typescript-eslint'

export default [
  {
    ignores: [
      'node_modules',
      'dist',
      'public',
      '.vscode',
      '.husky',
      '.git',
      'coverage',
      'eslint.config.js',
    ],
  },
  // Base JS
  js.configs.recommended,
  // TypeScript
  ...tseslint.configs.recommendedTypeChecked,
  ...tseslint.configs.stylisticTypeChecked,
  ...eslintPluginAstro.configs.recommended,
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.es2021,
      },
      parserOptions: {
        projectService: {
          allowDefaultProject: ['*.js', 'eslint.config.js'],
        },
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },
  // React
  reactPlugin.configs.flat.recommended,
  reactPlugin.configs.flat['jsx-runtime'],
  // React Hooks
  {
    plugins: {
      'react-hooks': reactHooksPlugin,
    },
    rules: {
      ...reactHooksPlugin.configs.recommended.rules,
    },
  },

  // A11y
  jsxA11yPlugin.flatConfigs.recommended,
  // Astro: отключаем type-checked правила — они запускают TS language service
  // для каждого .astro файла, что делает сохранение крайне медленным
  {
    files: ['**/*.astro'],
    ...tseslint.configs.disableTypeChecked,
  },
  // Astro overrides — после React-конфигов, чтобы не перекрывались
  {
    files: ['**/*.astro'],
    languageOptions: {
      parserOptions: {
        project: false,
      },
    },
    rules: {
      'react/no-unknown-property': 'off',
      'react/jsx-key': 'off',
      'react/react-in-jsx-scope': 'off',
      'jsx-a11y/anchor-is-valid': 'off',
    },
  },
  // Custom & Overrides
  {
    plugins: {
      'simple-import-sort': simpleImportSortPlugin,
      'only-warn': onlyWarn,
    },
    rules: {
      'simple-import-sort/imports': [
        'warn',
        {
          groups: [
            // node:.
            ['^node:'],
            // packages
            ['^react$', '^react-dom$', '^\\u0000', '^@?\\w'],
            // Project files
            ['^@/'],
            // ../
            ['^\\.\\.(?!/?$)', '^\\.\\./?$'],
            // ./
            ['^\\./(?=.*/)(?!/?$)', '^\\.(?!/?$)', '^\\./?$', '^.+\\.s?css$'],
          ],
        },
      ],
      'simple-import-sort/exports': 'warn',

      'react/prop-types': 'off',
      'react/display-name': 'off',
      'react/no-unknown-property': 'off',

      'no-console': ['warn', { allow: ['warn', 'error'] }],

      'no-param-reassign': ['warn', {
        props: true,
        ignorePropertyModificationsFor: ['acc', 'state'],
      }],

      'no-restricted-syntax': ['error', {
        selector: 'ForInStatement',
        message:
          'for..in loops iterate over the entire prototype chain, which is virtually never what you want. Use Object.{keys,values,entries}, and iterate over the resulting array.',
      }, {
        selector: 'LabeledStatement',
        message: 'Labels are a form of GOTO; using them makes code confusing and hard to maintain and understand.',
      }, {
        selector: 'WithStatement',
        message: '`with` is disallowed in strict mode because it makes code impossible to predict and optimize.',
      }],

      'no-underscore-dangle': ['warn', {
        allowAfterThis: true,
        allow: ['__filename', '__dirname'],
      }],

      '@typescript-eslint/no-unused-vars': ['warn', {
        argsIgnorePattern: '^_',
        varsIgnorePattern: '^_',
        caughtErrorsIgnorePattern: '^_',
      }],

      '@typescript-eslint/no-shadow': ['warn'],

      '@typescript-eslint/no-unused-expressions': ['warn', {
        allowShortCircuit: true,
        allowTernary: true,
        allowTaggedTemplates: true,
      }],

      '@typescript-eslint/consistent-type-definitions': 'off',
      '@typescript-eslint/consistent-indexed-object-style': 'off',
      '@typescript-eslint/no-redundant-type-constituents': 'off',
      '@typescript-eslint/no-misused-promises': 'off',
      '@typescript-eslint/non-nullable-type-assertion-style': 'off',
      '@typescript-eslint/no-extra-non-null-assertion': 'error',
      '@typescript-eslint/no-unsafe-member-access': 'off',
      '@typescript-eslint/no-unsafe-assignment': 'off',
      '@typescript-eslint/no-unsafe-argument': 'off',
      '@typescript-eslint/no-floating-promises': 'off',
      '@typescript-eslint/no-non-null-assertion': 'error',
      '@typescript-eslint/prefer-nullish-coalescing': 'off',

      // Отключаем правила форматирования дабы избежать конфликтов с dprint
      ...prettierConfig.rules,
    },
    settings: {
      react: {
        version: 'detect',
      },
    },
  },
]
