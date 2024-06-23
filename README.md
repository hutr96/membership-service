# Project 2
Timeout = 5 second
-h: hostfile
-t: testcase (only has 2, 3, and 4)
# Test cases
In proj2 folder: 

    docker build . -t proj2-test


**Test case 1:**
Run script (will stop and clean containers automatically after 10s). Logs are saved in ./logs folder.

    ./runproj2_case1.sh

Or manually run each process one by one. 

    docker run --rm -it --name host1 --network mynetwork -h host1 proj2-test -h hostfile
    docker run --rm -it --name host2 --network mynetwork -h host2 proj2-test -h hostfile
    docker run --rm -it --name host3 --network mynetwork -h host3 proj2-test -h hostfile
    docker run --rm -it --name host4 --network mynetwork -h host4 proj2-test -h hostfile


**Test case 2:**
Run script (will stop and clean containers automatically after 20s). Logs are saved in ./logs folder. It will start each peer and kill the last one after 5s. 

    ./runproj2_case2.sh

Or manually run each process. Then wait 5s, the last peer will crash (or kill any of them manually). 

    docker run --rm -it --name host1 --network mynetwork -h host1 proj2-test -h hostfile -t 2
    docker run --rm -it --name host2 --network mynetwork -h host2 proj2-test -h hostfile -t 2
    docker run --rm -it --name host3 --network mynetwork -h host3 proj2-test -h hostfile -t 2
    docker run --rm -it --name host4 --network mynetwork -h host4 proj2-test -h hostfile -t 2



**Test case 3:**
Run script (will stop and clean containers automatically after 45s). Logs are saved in ./logs folder. It will start each peer and then kill each peer except the leader every 5s. 

    ./runproj2_case3.sh

Or manually run each process. Then start crashing them one by one. Also should be able to handle two and more peers crashes at the same time. 

    docker run --rm -it --name host1 --network mynetwork -h host1 proj2-test -h hostfile
    docker run --rm -it --name host2 --network mynetwork -h host2 proj2-test -h hostfile
    docker run --rm -it --name host3 --network mynetwork -h host3 proj2-test -h hostfile
    docker run --rm -it --name host4 --network mynetwork -h host4 proj2-test -h hostfile



**Test case 4:**
Run script (will stop and clean containers automatically after 50s). Logs are saved in ./logs folder. In this case, after every peers join, the last peer (host4) will crash after 5s. Then, the current leader will also crash after he sent a request for removing that member. Note that the new leader (which is host2 in my testcase) has not received this delete information. The other peers (peer 3 in my case) will report the pending request to the new leader. After handling the pending request, the new leader will also remove the crashed old leader from the membership list. (Sometimes the script won't work, please run the processes manually if so)

    ./runproj2_case4.sh

Or manually run each process. Then crash one of the peer (non-leader), the current leader will automatically crash after he sent a request for removing that member. After handling the pending request, the new leader will also remove the crashed old leader from the membership list. 

    docker run --rm -it --name host1 --network mynetwork -h host1 proj2-test -h hostfile -t 4
    docker run --rm -it --name host2 --network mynetwork -h host2 proj2-test -h hostfile
    docker run --rm -it --name host3 --network mynetwork -h host3 proj2-test -h hostfile
    docker run --rm -it --name host4 --network mynetwork -h host4 proj2-test -h hostfile


