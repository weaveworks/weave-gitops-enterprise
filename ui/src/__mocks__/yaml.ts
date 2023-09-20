const yaml = jest.createMockFromModule<jest.Mock>('yaml');

function load() {
  return {};
}

yaml.load = load as jest.MockedFunction<typeof load>;

module.exports = yaml;
