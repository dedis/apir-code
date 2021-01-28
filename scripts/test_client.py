from subprocess import Popen, PIPE
import csv
import os
import time

# TODO: doesn't work with new client

# change dir to root of project
os.chdir("..")

# open test file
f = open('data/random_id_key.csv')
csv_f = csv.reader(f)

for row in csv_f:
    idUser = row[0]
    keyUser = row[1].strip()

    # run client
    start = time.time()
    process = Popen(["make", "run_client", "id="+idUser, "scheme=dpf"], stdout=PIPE)
    #process = Popen(["make", "run_client", "id="+idUser, "scheme=it"], stdout=PIPE)
    (output, err) = process.communicate()
    outputString = output.decode('utf-8')
    exit_code = process.wait()
    end = time.time()
    print(end - start)
    retrievedKey = (outputString.split("key: ",1)[1]).strip()
    if (retrievedKey != keyUser):
        print("ERROR")
