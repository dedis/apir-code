import random
import string
import secrets
import base64
import csv

def generate_id(length):
    allowed_chars = string.ascii_letters + '.' + '@' + '_' + '-'
    return ''.join(random.choice(allowed_chars) for _ in range(length))

def generate_key():
    return base64.b64encode(secrets.token_bytes(256))

def bytes_len(s):
    return len(s.encode('utf-8'))

filename = "random_id_key.csv"
fields = ["id", "key"]
n = 10000

with open(filename, 'w') as csvfile:
    csvwriter = csv.writer(csvfile)
    
    # write fileds
    csvwriter.writerow(fields)

    for x in range(n):
        row = [generate_id(32), generate_key().decode('utf-8')]
        csvwriter.writerow(row)

