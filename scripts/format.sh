# server peers file

serverFile="server.json"
fileName="productPeerConfig.json"
dst="/Users/anxin/Desktop/ghub-project/dkg-go/src/etc/"

echo '{' > $fileName
echo '"servers":[ ' >> $fileName


line=$(cat $serverFile | wc -l )
awk -v l="$line" '
{	
	if ( NR != l ) {
		print ($2":4000,")
	} else {
		print ($2":4000")
	}
}' $serverFile >> $fileName 

echo '] '>>$fileName
echo '} ' >>$fileName 
mv $fileName $dst$fileName


# ssh file

awk '{ 
	print ("ubuntu@"$2);
}' server.json  > log
rm server.json
mv log server.json
