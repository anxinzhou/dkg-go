from collections import defaultdict
import matplotlib.pyplot as plt
import os
from matplotlib.ticker import ScalarFormatter

nums=range(4,36,4)
orders=range(1,5)
dirName="processed_log"

# logs=[ "log/log"+str(num)+"_"+str(order) for num in nums for order in orders]
# print(logs)

statics=defaultdict(lambda :defaultdict(dict))
for num in nums:
	dic=defaultdict(list)
	for order in orders:
		file = os.path.join(dirName,"log"+str(num)+"_"+str(order))
		f=open(file)
		content=f.readlines()
		for c in content:
			c=c.strip('\n ms').split()
			name=' '.join(c[:-1])
			value=float(c[-1])
			dic[name].append(value)
			# if num==28 and name=="DKG setup":
			# 	print(value)
		f.close()
			
	for k in dic:
		v=dic[k]
		total=0
		count=0
		for value in v:
			if k=="Decryption" and (value>=8 or value<=1):
				continue
			elif k=="Encryption" and (value>=3):
				continue
			elif k=="DKG setup" and (value<=50):
				continue
			total+=value
			count+=1
		total=total/count/1000
		statics[k][num]=total

# begin draw
fig,ax = plt.subplots()
bar_width = 1.0
i=-(len(statics.keys())-2)/2
keys=["Encryption","Decryption","Combining Shares","DKG setup"]

for l in keys:
	# if l!="Decryption":
	# 	continue
	print(l)
	d=statics[l]
	x=sorted(d.keys())
	y=[d[k] for k in x]
	# print(y)
	if l=="DKG setup":
		print(statics[l])
		pass
		ax.plot(nums,y,marker='o',label=l)
	else:
		ax.bar([n+i*bar_width for n in nums],y,width=bar_width,align='center',alpha=0.5,label=l)
		i+=1
ax.set_yscale("log")
ax.get_yaxis().set_major_formatter(ScalarFormatter())
ax.set_xticks(nums)
ax.set_ylabel('Time (sec)')
ax.set_xlabel('Size of SM')
ax.legend(loc="upper right",bbox_to_anchor=(1.0, 0.85))
plt.show()




		