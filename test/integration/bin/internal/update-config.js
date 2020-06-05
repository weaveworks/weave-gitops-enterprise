import * as std from '@jkcfg/std';
import * as param from '@jkcfg/std/param';
import { merge } from '@jkcfg/std/merge';

const ITS_NOT_YAML = { encoding: std.Encoding.String };

const useLB = param.Boolean('useLB', false);
const numMasters = param.Number('numMasters', 1);
const configYamlPath = param.String('configYamlPath');
const numRequiredMachines = numMasters + (useLB ? 1 : 0) + 1;

function readHostsFile(path) {
  return std.read(path, ITS_NOT_YAML).then(data => {
    const lines = data.trim().split('\n');
    return lines.map(line => line.split(' '));
  });
}

function checkMachineCount(privateHosts, publicHosts) {
  if (
    privateHosts.length < numRequiredMachines ||
    publicHosts.length < numRequiredMachines
  ) {
    throw new Error(`
      The number of public addresses (${publicHosts.length}) or
      the number of private addresses (${privateHosts.length})
      doesn't match number of required machines (${numRequiredMachines}).
      (considering we need at least 1 worker.)
      `);
  }
}

function updateConfigYaml(config, machines, controlPlaneLbAddress = '') {
  return merge(config, {
    wksConfig: {
      controlPlaneLbAddress,
      sshConfig: {
        machines,
      },
    },
  });
}

function toMachines(privateHosts, publicHosts) {
  let controlPlaneLbAddress = '';
  let machines = privateHosts.map(([privateAddress], i) => {
    const [publicAddress] = publicHosts[i];
    return { publicAddress, privateAddress, role: 'worker' };
  });

  if (useLB) {
    [{ publicAddress: controlPlaneLbAddress }, ...machines] = machines;
  }

  // mutating, eh
  machines.slice(0, numMasters).forEach(machine => {
    machine.role = 'master';
  });

  return [machines, controlPlaneLbAddress];
}

function main() {
  return Promise.all([
    std.read(configYamlPath),
    readHostsFile('hosts_private'),
    readHostsFile('hosts_public'),
  ]).then(([config, privateHosts, publicHosts]) => {
    checkMachineCount(privateHosts, publicHosts);
    const [machines, controlPlaneLbAddress] = toMachines(
      privateHosts,
      publicHosts,
    );
    const newConfig = updateConfigYaml(config, machines, controlPlaneLbAddress);
    std.write(newConfig, configYamlPath);
  });
}

main();
