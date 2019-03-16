# server peers file

serverFile="server.json"
fileName="productPeerConfig.json"
dst="${HOME}/dkg-go/src/etc/"

echo '{' > $fileName
echo '"servers":[ ' >> $fileName

line=$(cat raw.json | wc -l )
awk -v l="$line" '
{	
	q="\""
	cmd="host "$2 " | awk '\''{ print $NF}'\''"
	cmd | getline s
	if ( NR != l ) {
		c= q""s":4000"q","
		print c
	} else {
		c= q""s":4000"q
		print c
	}
}' raw.json  >> $fileName 

echo '] '>>$fileName
echo '} ' >>$fileName 
mv $fileName $dst$fileName


# ssh file

awk '{ 
	print ("ubuntu@"$2);
}' raw.json  > server.json

