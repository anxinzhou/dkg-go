from collections import defaultdict
import matplotlib.pyplot as plt
import os
from matplotlib.ticker import ScalarFormatter

nums=range(4,36,4)
recordsNum=nums
orders=range(1,5)
dirName="processed_log"

# logs=[ "log/log"+str(num)+"_"+str(order) for num in nums for order in orders]
# print(logs)

def avg(l):
	return sum(l)/len(l)

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
			# if name=="Broadcast encryption":
			# 	print(value)
		f.close()
	if num not in recordsNum:
		continue
	for k in dic:
		v=dic[k]
		total=0
		count=0
		maxV=max(v)
		minV=min(v)
		for value in v:
			if k=="Decryption" and (value>=8 or value<=1):
				continue
			elif k=="Encryption" and (value>=3):
				continue
			elif k=="DKG setup" and (value<=50):
				continue
			elif k=="Broadcast encryption" and (value<=10):
				continue
			elif k=="Broadcast shares" and value>(maxV-minV)/2+minV:
				# print (maxV,minV)
				continue
			if k=="Broadcast encryption":
				print(value)
				pass
			total+=value
			count+=1
		total=total/count/1000
		# if k=="DKG setup":
		# 	statics[k][num]=maxV 
		# else:
		statics[k][num]=total

# begin draw
fig,ax = plt.subplots()
bar_width = 1.2
keys=["Encryption","Decryption","Combining Shares","DKG setup"]
colors=["orange","brown","navy","k"]
i=-(len(keys)-2)/2

print("Encryption:",avg(statics["Encryption"].values()))
print("Decryption:",avg(statics["Decryption"].values()))
print("Broadcast encryption:",avg(statics["Broadcast encryption"].values()))
print("Broadcast shares:",list(statics["Broadcast shares"].values())[-1])
print("Combining Shares:",list(statics["Combining Shares"].values())[-1])
print("DKG setup:",list(statics["DKG setup"].values()))
for l,c in zip(keys,colors):
	# if l!="Decryption":
	# 	continue
	d=statics[l]
	x=sorted(d.keys())
	y=[d[k] for k in x]
	# print(y)
	if l=="DKG setup":
		print(statics[l])
		pass
		ax.plot(recordsNum,y,marker='o',label=l,color=c)
	else:
		if l=="Combining Shares":
			b=ax.bar([n+i*bar_width for n in recordsNum],y,width=bar_width,align='center',label=l,color=c,hatch='//')
		else:
			b=ax.bar([n+i*bar_width for n in recordsNum],y,width=bar_width,align='center',label=l,color=c)		
		i+=1
ax.set_yscale("log")
ax.get_yaxis().set_major_formatter(ScalarFormatter())
ax.set_xticks(recordsNum)
ax.set_ylabel('Time (sec)',fontsize="14")
ax.set_xlabel('Size of secret-managing committee',fontsize="14")
ax.tick_params(labelsize="11")
ax.legend(loc="upper right",bbox_to_anchor=(1.0, 0.87),prop={"size":15})
plt.show()




		