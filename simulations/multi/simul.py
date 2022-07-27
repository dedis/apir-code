import multiprocessing
from multiprocessing import Process
import os
import tomli
from fabric import ThreadingGroup as Group

user = os.getenv('APIR_USER')
password = os.getenv('APIR_PASSWORD')
simul_dir = '/' + user + '/go/src/github.com/si-co/vpir-code/simulations/multi/'

default_server_command = "./server -id={} -scheme={} -dbLen={} -nRows={} -blockLen={}"

def test_command():
    return 'uname -a'

def load_config(file):
    with open(file, 'rb') as f:
        return tomli.load(f)

def load_servers_config():
    return load_config('config.toml')

def load_general_config():
    return load_config('simul.toml')

def load_individual_config(config_file):
    return load_config(config_file)

def get_servers_addresses():
    addrs = []
    config = load_servers_config()
    for s in config["servers"].values():
        addrs.append(s['ip'])

    return addrs

def servers_pool():
    servers = get_servers_addresses()
    return Group(*servers, 
            user=user, 
            connect_kwargs={'password': password,},
            )

def server_setup(c):
    with c.cd(simul_dir), c.prefix('PATH=$PATH:/usr/local/go/bin'):
        c.run('bash ' + 'setup.sh')

def server_pir_classic(c, sid, dbLen, nRows, blockLen):
    with c.cd(simul_dir + 'server'):
        print(default_server_command.format(sid, 'pir-classic', dbLen, nRows, blockLen))
        c.run(default_server_command.format(sid, 'pir-classic', dbLen, nRows, blockLen))

def server_pir_merkle(c, sid, dbLen, nRows, blockLen):
    with c.cd(simul_dir + 'server'):
       c.run(default_server_command.format(sid, 'pir-merkle', dbLen, nRows, blockLen))

def experiment_pir_classic(p):
    gc = load_general_config()
    ic = load_individual_config('pirClassic.toml')
    first = [p[0], 0, gc['DBBitLengths'][0], ic['NumRows'], ic['BlockLength']]
    second = [p[1], 1, gc['DBBitLengths'][0], ic['NumRows'], ic['BlockLength']]
    arguments = [first, second]
    ppool = multiprocessing.Pool(2)
    ppool.map(server_pir_classic, arguments)
    # for i, c in enumerate(p):
        # print(i)
        # server_pir_classic(c, 0, gc['DBBitLengths'][0], ic['NumRows'], ic['BlockLength'])

pool = servers_pool()
for c in pool:
    server_setup(c)
experiment_pir_classic(pool)
