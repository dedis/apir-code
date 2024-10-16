import argparse
import time
import requests
import multiprocessing
from multiprocessing import Process
import os
import tomli
from fabric import ThreadingGroup as Group
from fabric import Connection

# servers data
# user = os.getenv('APIR_USER')
# password = os.getenv('APIR_PASSWORD')
# path = os.getenv('APIR_PATH')
user = "root"
password = "dedislab"
path = "/root/go/src/github.com/si-co/vpir-code"

# simulations directory on servers
simul_dir = path + '/simulations/multi/'

# commands 
default_pir_server_command = "screen -dm ./server -logFile={} -scheme={} -dbLen={} -elemBitSize={} -nRows={} -blockLen={} -servers={} && sleep 15"
default_fss_server_command = "screen -dm ./server -logFile={} -scheme={} && sleep 15"
default_pir_client_command = "./client -logFile={} -scheme={} -repetitions={} -elemBitSize={} -bitsToRetrieve={}"
default_pir_client_multi_command = "./client -logFile={} -scheme={} -repetitions={} -elemBitSize={} -bitsToRetrieve={} -numServers={}"
default_fss_client_command = "./client -logFile={} -scheme={} -repetitions={} -inputSize={}"

# local directories
results_dir = "results_vectors"

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

def servers_addresses():
    addrs = []
    config = load_servers_config()
    for s in config["servers"].values():
        addrs.append(s['ip'])

    return addrs

def kill_servers(servers):
    for s in servers:
        requests.get("http://" + s + ":8080")

def client_address():
    config = load_servers_config()
    return config["client"]

def servers_pool():
    servers = servers_addresses()
    return Group(*servers, 
            user=user, 
            connect_kwargs={'password': password,},
            )

def two_servers_pool():
    servers = servers_addresses()
    two_servers = servers[0:2]
    return Group(*two_servers, 
            user=user, 
            connect_kwargs={'password': password,},
            )

def server_setup(c, sid):
    # enable agent forwarding for git pull
    c.forward_agent = True
    with c.cd(simul_dir), c.prefix('PATH=$PATH:/usr/local/go/bin'):
        #c.run('git pull', warn = True)
        c.run('echo '+ str(sid) + ' > server/sid')
        c.run('bash ' + 'setup.sh')
    # disable now useless agent forwarding
    c.forward_agent = False
    
    # upload config
    c.put('config.toml', remote=simul_dir)

def client_setup(c):
    # enable agent forwarding for git pull
    c.forward_agent = True
    with c.cd(simul_dir), c.prefix('PATH=$PATH:/usr/local/go/bin'):
        #c.run('git pull', warn = True)
        c.run('bash ' + 'setup.sh')
    # disable now useless agent forwarding
    c.forward_agent = False

    # upload config
    c.put('config.toml', remote=simul_dir)

def server_pir_command(logFile, scheme, dbLen, elemBitSize, nRows, blockLen, servers=2):
    return default_pir_server_command.format(logFile, scheme, dbLen, elemBitSize, nRows, blockLen, servers)

def client_pir_command(logFile, scheme, repetitions, elemBitSize, bitsToRetrieve):
    return default_pir_client_command.format(logFile, scheme, repetitions, elemBitSize, bitsToRetrieve)

def client_pir_multi_command(logFile, scheme, repetitions, elemBitSize, bitsToRetrieve, numServers):
    return default_pir_client_multi_command.format(logFile, scheme, repetitions, elemBitSize, bitsToRetrieve, numServers)

def server_fss_command(logFile, scheme):
    return default_fss_server_command.format(logFile, scheme)

def client_fss_command(logFile, scheme, repetitions, inputSize):
    return default_fss_client_command.format(logFile, scheme, repetitions, inputSize)

def experiment_pir(pir_type, server_pool, client):
    print('Experiment PIR', pir_type)
    gc = load_general_config()
    ic = load_individual_config('pir_' + pir_type + '.toml')

    print("\t Run", len(server_pool), "servers")
    # define experiment parameters
    databaseLengths = gc['DBBitLengths']
    rep = gc['Repetitions']
    ebs = ic['ElementBitSize']
    nr = ic['NumRows']
    bl = ic['BlockLength']
    btr = gc['BitsToRetrieve']

    # run experiment on all database lengths
    for dl in databaseLengths:
        logFile = "pir_" + pir_type + "_" + str(dl) + ".log"
        print("\t Starting", len(server_pool), "servers with database length", dl, "element bit size", ebs, "number of rows", nr, "block length", bl)
        print("\t server command:", server_pir_command(logFile, "pir-" + pir_type, dl, ebs, nr, bl))
        server_pool.run('cd ' + simul_dir + 'server && ' + server_pir_command(logFile, "pir-" + pir_type, dl, ebs, nr, bl))
        if "classic" in pir_type:
            time.sleep(30)
        else:
            time.sleep(900)
        print("\t Run client")
        client.run('cd ' + simul_dir + 'client && ' + client_pir_command(logFile, "pir-" + pir_type, rep, ebs, btr))

        kill_servers(servers_addresses())

    # get all log files
    for dl in databaseLengths:
        logFile = "pir_" + pir_type + "_" + str(dl) + ".log"
        for i, c in enumerate(server_pool):
            print("\t server", str(i), "log file location:", simul_dir + 'server/' + logFile)
            c.get(simul_dir + 'server/' + logFile, results_dir + "/server_" + str(i) + "_" + logFile)

        print("\t client", "log file location:", simul_dir + 'client/' + logFile)
        client.get(simul_dir + 'client/' + logFile, results_dir + "/client_" + logFile)

def experiment_pir_vector(pir_type, server_pool, client):
    print('Experiment PIR vector', pir_type)
    gc = load_general_config()
    ic = load_individual_config('pir_' + pir_type + '_vector.toml')

    print("\t Run", len(server_pool), "servers")
    # define experiment parameters
    databaseLengths = gc['DBBitLengths']
    rep = gc['Repetitions']
    ebs = ic['ElementBitSize']
    nr = ic['NumRows']
    bl = ic['BlockLength']
    btr = gc['BitsToRetrieve']

    # run experiment on all database lengths
    for dl in databaseLengths:
        logFile = "pir_" + pir_type + "_vector_" + str(dl) + ".log"
        print("\t Starting", len(server_pool), "servers with database length", dl, "element bit size", ebs, "number of rows", nr, "block length", bl)
        print("\t server command:", server_pir_command(logFile, "pir-" + pir_type, dl, ebs, nr, bl))
        server_pool.run('cd ' + simul_dir + 'server && ' + server_pir_command(logFile, "pir-" + pir_type, dl, ebs, nr, bl))
        if "classic" in pir_type:
            time.sleep(30)
        else:
            time.sleep(900)
        print("\t Run client")
        client.run('cd ' + simul_dir + 'client && ' + client_pir_command(logFile, "pir-" + pir_type, rep, ebs, btr))

        kill_servers(servers_addresses())

    # get all log files
    for dl in databaseLengths:
        logFile = "pir_" + pir_type + "_" + str(dl) + ".log"
        for i, c in enumerate(server_pool):
            print("\t server", str(i), "log file location:", simul_dir + 'server/' + logFile)
            c.get(simul_dir + 'server/' + logFile, results_dir + "/server_" + str(i) + "_" + logFile)

        print("\t client", "log file location:", simul_dir + 'client/' + logFile)
        client.get(simul_dir + 'client/' + logFile, results_dir + "/client_" + logFile)

def experiment_pir_multi(pir_type, server_pool, client):
    print('Experiment PIR multi', pir_type)
    gc = load_general_config()
    ic = load_individual_config('pir_' + pir_type + '_multi.toml')

    # define experiment parameters
    dl = 8589935000 # 1 GiB for this experiment
    rep = gc['Repetitions']
    ebs = ic['ElementBitSize']
    nr = ic['NumRows']
    bl = ic['BlockLength']
    btr = gc['BitsToRetrieve']
    numServers = ic['NumServers']

    # run experiment on all database lengths
    for s in numServers:
        logFile = "pir_" + pir_type + "_multi_" + str(s) + ".log"
        print("\t Starting", str(s), "servers with database length", dl, "element bit size", ebs, "number of rows", nr, "block length", bl)
        print("\t server command:", server_pir_command(logFile, "pir-" + pir_type, dl, ebs, nr, bl, s))
        server_pool.run('cd ' + simul_dir + 'server && ' + server_pir_command(logFile, "pir-" + pir_type, dl, ebs, nr, bl, s))
        time.sleep(100)
        print("\t Run client")
        print("\t client command:", client_pir_multi_command(logFile, "pir-" + pir_type, rep, ebs, btr, s))
        client.run('cd ' + simul_dir + 'client && ' + client_pir_multi_command(logFile, "pir-" + pir_type, rep, ebs, btr, s))

        kill_servers(servers_addresses())

    # get all log files
    for s in numServers:
        logFile = "pir_" + pir_type + "_multi_" + str(s) + ".log"
        for i, c in enumerate(server_pool):
            print("\t server", str(i), "log file location:", simul_dir + 'server/' + logFile)
            c.get(simul_dir + 'server/' + logFile, results_dir + "/server_" + str(i) + "_" + logFile)

        print("\t client", "log file location:", simul_dir + 'client/' + logFile)
        client.get(simul_dir + 'client/' + logFile, results_dir + "/client_" + logFile)

        # delete useless logs from servers that weren't used in the specific iteration
        directory = os.fsencode(results_dir)
        for file in os.listdir(directory):
            filename = os.fsdecode(file)
            if "multi" in filename and 'stats' not in open(results_dir + "/" + filename).read():
                print("removing uselss log file", filename)
                os.remove(results_dir + "/" + filename)

def experiment_pir_classic(server_pool, client):
    experiment_pir("classic", server_pool, client)

def experiment_pir_merkle(server_pool, client):
    experiment_pir("merkle", server_pool, client)

def experiment_pir_classic_vector(server_pool, client):
    experiment_pir_vector("classic", server_pool, client)

def experiment_pir_merkle_vector(server_pool, client):
    experiment_pir_vector("merkle", server_pool, client)

def experiment_pir_multi_classic(server_pool, client):
    experiment_pir_multi("classic", server_pool, client)

def experiment_pir_multi_merkle(server_pool, client):
    experiment_pir_multi("merkle", server_pool, client)

def experiment_fss(fss_type, server_pool, client):
    print('Experiment FSS', fss_type)
    gc = load_general_config()
    ic = load_individual_config('fss_' + fss_type + '.toml')
    print("\t Run", len(server_pool), "servers")
    # define experiment parameters
    rep = gc['Repetitions']
    inputSizes = ic['InputSizes']

    # run experiment on all input sizes
    for inputSize in inputSizes:
        logFile = "fss_" + fss_type + "_" + str(inputSize) + ".log"
        print("\t server command:", server_fss_command(logFile, "fss-" + fss_type))
        server_pool.run('cd ' + simul_dir + 'server && ' + server_fss_command(logFile, "fss-" + fss_type))
        time.sleep(30)
        print("\t Run client")
        client.run('cd ' + simul_dir + 'client && ' + client_fss_command(logFile, "fss-" + fss_type, rep, inputSize))

        kill_servers(servers_addresses())

    # get all log files
    for inputSize in inputSizes:
        logFile = "fss_" + fss_type + "_" + str(inputSize) + ".log"
        for i, c in enumerate(server_pool):
            print("\t server", str(i), "log file location:", simul_dir + 'server/' + logFile)
            c.get(simul_dir + 'server/' + logFile, results_dir + "/server_" + str(i) + "_" + logFile)

        print("\t client", "log file location:", simul_dir + 'client/' + logFile)
        client.get(simul_dir + 'client/' + logFile, results_dir + "/client_" + logFile)

def experiment_fss_classic(server_pool, client):
    experiment_fss("classic", server_pool, client)

def experiment_fss_auth(server_pool, client):
    experiment_fss("auth", server_pool, client)

# Setup server and client
print("Servers' setup")
pool = servers_pool()
for i, c in enumerate(pool):
    print("\t Setting up server", i, "with Fabric connection", c)
    server_setup(c, i)
print("Client's setup")
client_host = client_address()
client = Connection(client_host, user=user, connect_kwargs={'password': password,})
client_setup(client)

# -----------Argument Parser-------------
parser = argparse.ArgumentParser()
parser.add_argument(
    "-e",
    "--expr",
    type=str,
    help="experiments: point, point_multi, predicate",
    required=True)

args = parser.parse_args()
EXPR = args.expr

if __name__ == "__main__":
    if not os.path.exists("figures"):
        os.makedirs("figures")
    if not os.path.exists("results"):
        os.makedirs("results")

    if EXPR == "point":
        # run experiments, 
        # in this case only with two servers
        experiment_pir_classic(pool, client)
        experiment_pir_merkle(pool, client)
    if EXPR == "vector":
        # run experiments, 
        # in this case only with two servers
        experiment_pir_classic_vector(pool, client)
        experiment_pir_merkle_vector(pool, client)
    elif EXPR == "point_multi":
        # run multi experiments, 
        # with all the servers
        experiment_pir_multi_classic(pool, client)
        experiment_pir_multi_merkle(pool, client)
    elif EXPR == "predicate":
        # run experiments for 
        # fss with only two servers
        experiment_fss_classic(pool, client)
        experiment_fss_auth(pool, client)
    else:
        print("Unknown experiment: choose between the available options")




