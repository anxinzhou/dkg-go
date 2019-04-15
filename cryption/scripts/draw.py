from collections import defaultdict
import matplotlib.pyplot as plt
import os
from matplotlib.ticker import ScalarFormatter

nums = range(4, 36, 4)
dirName = "log"


# logs=[ "log/log"+str(num)+"_"+str(order) for num in nums for order in orders]
# print(logs)

def avg(l):
    return sum(l) / len(l)


keys = ["Encryption","Combining shares"]
statics = defaultdict(lambda: defaultdict(dict))
for num in nums:
    dic = defaultdict()
    file = os.path.join(dirName, "log" + str(num))
    f = open(file)
    content = f.readlines()
    v1 = content[1].strip('\n ms').split()[-1]
    v2 = content[2].strip('\n ms').split()[-1]
    statics[keys[0]][num] = v1
    statics[keys[1]][num] = v2
    f.close()

# begin draw
fig, ax = plt.subplots()
bar_width = 1.2
colors = ["#EFDC05", "#E53A40"]
i = -(len(keys) - 2) / 2

print(statics)
for l, c in zip(keys, colors):
    # if l!="Decryption":
    # 	continue
    d = statics[l]
    x = sorted(d.keys())
    y = [float(d[k]) for k in x]
    # print(y)
    if l == "Combining shares":
        b = ax.bar([n + i * bar_width for n in nums], y, width=bar_width, align='center', label=l, color=c,
                   hatch='//')
    else:
        b = ax.bar([n + i * bar_width for n in nums], y, width=bar_width, align='center', label=l, color=c)
    i += 1
ax.set_xticks(nums)
# ax.set_yticks([])
ax.set_ylabel('Time (ms)', fontsize="14")
ax.set_xlabel('Size of secret-managing committee', fontsize="14")
# ax.tick_params(labelsize="11")
# ax.set_yticks(np.arange(0,28,4))
# labelspacing
ax.get_yaxis().set_major_formatter(ScalarFormatter())
ax.legend(loc='upper left',prop={"size":16},handletextpad=2)
plt.yticks(range(0,44,4))
# ax.legend(loc='center left', prop={"size": 30}, handletextpad=4.2, labelspacing=1.2)
plt.show()
