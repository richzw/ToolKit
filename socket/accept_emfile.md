Q: When a process runs out of file descriptors, `accept()` will fail and set errno to EMFILE
---------------

You can set aside an extra fd at the beginning of your program and keep track of the `EMFILE` condition:

```c
int reserve_fd;
_Bool out_of_fd = 0;

if(0>(reserve_fd = dup(1)))
err("dup()");
```

Then, if you hit the `EMFILE` condition, you can close the `reserve_fd` and use its slot to accept the new connection (which you'll then immediately close):

```c
clientfd = accept(serversocket,(struct sockaddr*)&client_addr,&client_len);
if (out_of_fd){
close(clientfd);
if(0>(reserve_fd = dup(1)))
err("dup()");
out_of_fd=0;

    continue; /*doing other stuff that'll hopefully free the fd*/
}

if(clientfd < 0)  {
close(reserve_fd);
out_of_fd=1;
continue;
}
```

Complete Ex

```cpp
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <errno.h>
#include <string.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>


static void err(const char *str)
{
    perror(str);
    exit(1);
}


int main(int argc,char *argv[])
{
    int serversocket;
    struct sockaddr_in serv_addr;
    serversocket = socket(AF_INET,SOCK_STREAM,0);
    if(serversocket < 0)
        err("socket()");
    int yes;
    if ( -1 == setsockopt(serversocket, SOL_SOCKET, SO_REUSEADDR, &yes, sizeof(int)) )
        perror("setsockopt");


    memset(&serv_addr,0,sizeof serv_addr);

    serv_addr.sin_family = AF_INET;
    serv_addr.sin_addr.s_addr= INADDR_ANY;
    serv_addr.sin_port = htons(6543);
    if(bind(serversocket,(struct sockaddr*)&serv_addr,sizeof serv_addr) < 0)
        err("bind()");

    if(listen(serversocket,10) < 0)
        err("listen()");

    int reserve_fd;
    int out_of_fd = 0;

    if(0>(reserve_fd = dup(1)))
        err("dup()");


    for(;;) {
        struct sockaddr_storage client_addr;
        socklen_t client_len = sizeof client_addr;
        int clientfd;


        clientfd = accept(serversocket,(struct sockaddr*)&client_addr,&client_len);
        if (out_of_fd){
            close(clientfd);
            if(0>(reserve_fd = dup(1)))
                err("dup()");
            out_of_fd=0;

            continue; /*doing other stuff that'll hopefully free the fd*/
        }

        if(clientfd < 0)  {
            close(reserve_fd);
            out_of_fd=1;
            continue;
        }

    }

    return 0;
}
```

makes sense if 
- 1) the listening socket isn't shared with other processes (which might not have hit their EMFILE limit yet) 
- 2) the server deals with persistent connections (because if it doesn't, then you're bound to close some existing connection very soon, freeing up a fd slot for your next attempt at accept).
