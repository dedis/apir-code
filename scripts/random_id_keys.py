import random
import string
import secrets
import base64
import csv
import os

def generate_id(length):
    allowed_chars = string.ascii_letters + '.' + '@' + '_' + '-'
    return ''.join(random.choice(allowed_chars) for _ in range(length))

def generate_key():
    return base64.b64encode(secrets.token_bytes(258))

def bytes_len(s):
    return len(s.encode('utf-8'))

filename = "../data/random_id_key.csv"
n = 10000

# remove file first
os.remove(filename)

with open(filename, 'w') as csvfile:
    csvwriter = csv.writer(csvfile)
    
    for x in range(n):
        row = [generate_id(32), generate_key().decode('utf-8')]
        csvwriter.writerow(row)
