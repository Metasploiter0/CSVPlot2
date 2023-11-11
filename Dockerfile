FROM debian

RUN apt-get update && apt-get upgrade 
RUN apt-get install curl
RUN curl https://go.dev/dl/go1.21.4.linux-amd64.tar.gz
RUN rm -rf /usr/local/go && tar -C /usr/local -xzf go1.21.4.linux-amd64.tar.gz
# RUN rm -rf /usr/local/go && tar -C /usr/local -xzf https://go.dev/dl/go1.21.4.linux-amd64.tar.gz
RUN export PATH=$PATH:/usr/local/go/bin
RUN go version 

CMD ["echo", "Containered version of debian is now set up for usage"]
