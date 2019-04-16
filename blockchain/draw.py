from collections import defaultdict
import matplotlib.pyplot as plt
import os
from matplotlib.ticker import ScalarFormatter
from scipy.stats import geom
import numpy as np

fig, ax = plt.subplots()
lenOfHash = 256
hashPow = 2200000*60    # hash number per min
lam = [227,228,229]      # larger lam means larger difficulty
committeeNumber = 32
pointNumber = 2<<8
minX=1
maxX=15
x = np.linspace(minX,maxX,pointNumber)   # x from 1 min to 10 min
colors = ["#EFDC05","#30A9DE","#E53A40"]
# 227
for lm,c in zip(lam,colors):
    p = 1/(2<<(lenOfHash-lm))
    y = geom.cdf(x*hashPow, p) * committeeNumber
    print(geom.cdf(x*hashPow,p))
    ax.plot(x, y, label="\u03BB="+str(lm),c=c)


ax.set_ylabel('Expected number', fontsize="12")
ax.set_xlabel('Time (min)', fontsize="12")
ax.legend(loc='lower right',prop={"size":16},handletextpad=1)
plt.show()
