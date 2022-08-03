import time
import requests
import multiprocessing
from multiprocessing import Process
import os
import tomli
from fabric import ThreadingGroup as Group
from fabric import Connection

user = os.getenv('APIR_USER')
password = os.getenv('APIR_PASSWORD')
simul_dir = '/' + user + '/go/src/github.com/si-co/vpir-code/simulations/multi/'
default_pir_server_command = "screen -dm ./server -logFile={} -scheme={} -dbLen={} -elemBitSize={} -nRows={} -blockLen={} && sleep 15"
default_pir_client_command = "./client -logFile={} -scheme={} -repetitions={} -elemBitSize={} -bitsToRetrieve={}"
default_pir_client_multi_command = "./client -logFile={} -scheme={} -repetitions={} -elemBitSize={} -bitsToRetrieve={} -numServers={}"
results_dir = "results"

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

def client_address():
    config = load_servers_config()
    return config["client"]

def servers_pool():
    servers = servers_addresses()
    return Group(*servers, 
            user=user, 
            connect_kwargs={'password': password,},
            )

def server_setup(c, sid):
    # enable agent forwarding for git pull
    c.forward_agent = True
    with c.cd(simul_dir), c.prefix('PATH=$PATH:/usr/local/go/bin'):
        c.run('git pull', warn = True)
        c.run('echo '+ str(sid) + ' > server/sid')
        c.run('bash ' + 'setup.sh')
    # disable now useless agent forwarding
    c.forward_agent = False

def client_setup(c):
    # enable agent forwarding for git pull
    c.forward_agent = True
    with c.cd(simul_dir), c.prefix('PATH=$PATH:/usr/local/go/bin'):
        c.run('git pull', warn = True)
        c.run('bash ' + 'setup.sh')
    # disable now useless agent forwarding
    c.forward_agent = False

def server_pir_command(logFile, scheme, dbLen, elemBitSize, nRows, blockLen):
    return default_pir_server_command.format(logFile, scheme, dbLen, elemBitSize, nRows, blockLen)

def client_pir_command(logFile, scheme, repetitions, elemBitSize, bitsToRetrieve):
    return default_pir_client_command.format(logFile, scheme, repetitions, elemBitSize, bitsToRetrieve)

def client_pir_multi_command(logFile, scheme, repetitions, elemBitSize, bitsToRetrieve, numServers):
    return default_pir_client_multi_command.format(logFile, scheme, repetitions, elemBitSize, bitsToRetrieve, numServers)

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
        time.sleep(300)
        print("\t Run client")
        client.run('cd ' + simul_dir + 'client && ' + client_pir_command(logFile, "pir-" + pir_type, rep, ebs, btr))
        # kill servers
        for s in servers_addresses():
            requests.get("http://" + s + ":8080")

    # get all log files
    for dl in databaseLengths:
        logFile = "pir_" + pir_type + "_" + str(dl) + ".log"
        for i, c in enumerate(server_pool):
            print("\t server", str(i), "log file location:", simul_dir + 'server/' + logFile)
            c.get(simul_dir + 'server/' + logFile, results_dir + "/server_" + str(i) + "_" + logFile)

        print("\t client", "log file location:", simul_dir + 'client/' + logFile)
        client.get(simul_dir + 'client/' + logFile, results_dir + "/client_" + logFile)

def experiment_pir_multi(pir_type, server_pool, client):
    # print('Experiment PIR multi', pir_type)
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
        print("\t server command:", server_pir_command(logFile, "pir-" + pir_type, dl, ebs, nr, bl))
        server_pool.run('cd ' + simul_dir + 'server && ' + server_pir_command(logFile, "pir-" + pir_type, dl, ebs, nr, bl))
        time.sleep(30)
        print("\t Run client")
        print("\t client command:", client_pir_multi_command(logFile, "pir-" + pir_type, rep, ebs, btr, s))
        client.run('cd ' + simul_dir + 'client && ' + client_pir_multi_command(logFile, "pir-" + pir_type, rep, ebs, btr, s))
        # kill servers
        for s in servers_addresses():
            requests.get("http://" + s + ":8080")

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

def experiment_pir_multi_classic(server_pool, client):
    experiment_pir_multi("classic", server_pool, client)

def experiment_pir_multi_merkle(server_pool, client):
    experiment_pir_multi("merkle", server_pool, client)

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

# run experiments, in this case only with two servers
#experiment_pir_classic(pool[0:2], client)
#experiment_pir_merkle(pool[0:2], client)

# run multi experiments, with all the servers
#experiment_pir_multi_classic(pool, client)
#experiment_pir_multi_merkle(pool, client)
