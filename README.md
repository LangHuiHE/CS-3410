# Distributed System Models in Go

## peer-to-peer | concurrency > MUD
MUD is a kind of multi-player text role-playing game that eventually evolved into the MMORPGs of today. In order to support the multi-player environment which, include concurrency, networking, and multiple players who can interact, the program sets up different goroutines, which are lightweight thread, on master servicer.

## containers > Chord
DHT implements the services to maintain a DHT, and it supports a simple command-line user interface. Each instance (running on different machines, or the same machine on different ports) will run an RPC server and act as an RPC client. Besides, several background functions will run to keep the DHT data structures up-to-date.

## consensus > Paxos
A simulator that implements the protocol, with a few key actions being controlled by a script. This program was based on the idea of the paper called Paxos Made Simple and finite-state machine (FSM) behavior model.

## databases > MapReduce
This infrastructure includes a master node and worker nodes connected by the PRC method. Word counting is used for the test case in this project. The numbers of the tasks of mapping and reducing are flexible The source database gets split into the number of mapping tasks. Works use HTTP to download the file from the master for mapping. They send the file’s addresses back to the master when mapping is completed. Then the master sends out reduce tasks with the file’s addressed to workers. After the master downloads and merges the reduced files from works, the master tells the worker to clean up and shut down then the master shuts down finishing the whole program.
