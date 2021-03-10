import matplotlib.pyplot as plt
import numpy as np
import csv

workers = []
time = []
with open('results.csv', 'r') as csv_file:
    reader = csv.reader(csv_file)
    for row in reader:
        workers.append(row[0])
        time.append(int(row[1]))

fig, ax = plt.subplots()

# Example data
#x_pos = np.arange(len(workers))
ax.bar(workers,time)

# ax.barh(y_pos, time, align='center')
ax.set_xticks(workers)
# ax.set_yticklabels(workers)
# ax.invert_yaxis()  # labels read top-to-bottom
# ax.set_xlabel('Wall time [ms]')

plt.show()
