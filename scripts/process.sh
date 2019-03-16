#/usr/bin/bash

srcDir="log"
dstDir="processed_log"


for ((i=4;i<=32;i+=4))
do
	for ((j=1;j<=4;j+=1))
	do
		file="log${i}_${j}"
		src="${srcDir}/${file}"
		dst="${dstDir}/${file}"
		awk '/time/ { $1="";$2=""; $3=""; print $0}' $src  | grep -v "dkg.go" |  sed -e 's/[!:]//g'  -e 's/^ *//g' -e '/wait/d' -e 's/time//g' \
		-e '/stage./d' \
		-e '/receiving encrption/d' \
		-e 's/combine share/Combining Shares/g' -e 's/decrption/Decryption/g' \
		-e 's/encrytion/Encryption/g' -e 's/total dkg/DKG setup/g' \
		-e '/receiving share/d' -e '/decryption total/d' | sort > $dst
	done
done