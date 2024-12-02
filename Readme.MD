Ok So lets see how this distributed file storage system works

What it does?
It is used to store files as in network for organization
Like an organization can have a port number for their company and they can store their data to this network and if they want to get the data of other organization they can ask for their encryption key and can decode their data

In this every network have data of every connected network to it but it is decoded and only the organization having the encryption key can view it. For example in this we are storing the data for the server3 but its data is available in server1 and server2

The file name we pass will be encoded to a path and the file will be stored at that path
And the file name will be the path to that file so it is highly encrypted



Before starting the server we must first understand some components like transport, message, server, store, tcp, handshake, crypto and timeSleep

1. Transport
It is way to communicate between clients(peer) between networks or within the network. In this we have created the tcp transport protocol

2. Message
Message is the data that is shared between the peers

3. Server
It is where all the peers communicate and its where we have created the storage of different networks and for different peers

4. Store
It is where we store all the data of different networks and peer

5. tcp
It is type of transport in which no data lose is there

6. handshake
It is when two networks are connected to each other and sending and recieving of data happens

7. crypto
It is used to encrypt and decrypt the data so that even if a networks data is stored on other network yet the other netwoek can't read the data but the data to which it belongs can easily decrypt it and read it 

8. timeSleep
It is used to wait so that we give time to create a server before sending it is as a peer



We start with creating a server using the makeserver command in the main.go
In make server command we first specify the port number and then whom it is connected to port number
In the make server function we first create a tcp transport and a file server where we can store the data
tcpTransportOpts means options available in tcp transport like listenaddr(port number of the server we are creating) then the nodes(the server we want to connect with)

After we create the tcp transport we create the server in which we first specify the encryption key that in future help to retrieve data from the other network when we lost the data on self network
Then we create the storage root(this means the folder name of root)
After we pass the pathTransformFunc(This will help to decode the filename and create a system in which files are stored in multilevel heirarchial structure i.e. folder under folder)
Then we specify the tcp transport
We then specify which node we want to get connected i.e. other servers

This will return the server instance
Here in the example we have created three server in which server 1 is not connected to any server beacuse no server was created before it. 2 server is connected to 1 server and the 3 server is connected to server 1 and 2.
After we get their instances the s1,s2,s3 we have to start them
Here if we just start them normally then it will block the rest of the code of main so we have to make them run parallely so we have sent them in go func. If we sent start of two server in a single go func then it would result in starting of first server only because it will block after to listen and write

After we start the server in which it calls the function ListenandAccept which then calls the net.Listen which starts the port on local host with the port number we have mentioned
In the listenAndAccept only we start the startAcceptLoop which reads the message we want to store in the files
This will make a infinte while loop that can only be ended if we close the connection
We then send the connection to the handleConn function
This will create a newPeer(a new client) and we create a new handshake between the peer and server

In main after we have created the server and connected them to each other we store the data in server 3 and watch what will happen.The data stored in server 3 will be stored in each server 1 and 2 but it will be encrypted and can't be read by any other server
Only the server3 can read it beacuse each server has its own encryption key and it helps to decode the file contents
So if by any chance the data of server3 is deleted it can send message to the server that it needs its data
The server will look for the encryption key of server3 in other servers and then return the data stored under that key
