import math

sigma = 6.4
secparam = 128
db_sizes = [1024, 1048576, 1073741824] # in bits

bounds = []

for dbs in db_sizes:
    bounds.append(math.ceil(math.sqrt(dbs * secparam) * sigma))

q = pow(2, 32)

epsilons = []
for B in bounds:
    epsilons.append((2*B - 1)/(q - 4*B + 1))

print("integrity error 2^-128")
for i, epsilon in enumerate(epsilons):
    for t in range(1, 15):
        if pow(epsilon, t + 1) < 1/pow(2, 128):
            print("\t DB size:", db_sizes[i], "B: ", bounds[i], "t: ", t)
            break

print("integrity error 2^-64")
for i, epsilon in enumerate(epsilons):
    for t in range(1, 15):
        if pow(epsilon, t + 1) < 1/pow(2, 64):
            print("\t DB size:", db_sizes[i], "B: ", bounds[i], "t: ", t)
            break
