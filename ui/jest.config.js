module.exports = {
  preset: '@vue/cli-plugin-unit-jest',
  setupFiles: ['./tests/index.js'],
  verbose: true,
  moduleNameMapper: {
    '\\.(css|less|sass|scss)$': '<rootDir>/tests/__mocks__/styleMock.js',
    '\\.(gif|ttf|eot|svg)$': '<rootDir>/tests/__mocks__/fileMock.js',
  },
};
