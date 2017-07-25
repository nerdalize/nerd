FROM ubuntu:16.04

RUN apt-get update && apt-get install -y \
	vim \
	curl \
	perl \
	cpanminus \
	perl-doc


RUN mkdir -p /library/yzhang/PSSpred/

RUN cd /library/yzhang/PSSpred && curl http://zhanglab.ccmb.med.umich.edu/PSSpred/PSSpred_v3.tar.gz | tar xz 

# This is a 4GB download!
RUN cd /library/yzhang/ && curl ftp://zhanglab.ccmb.med.umich.edu/download/nr.tar.gz | tar xz

RUN cd /library/yzhang/ && curl http://zhanglab.ccmb.med.umich.edu/PSSpred/blastv2.6.tar.gz | tar xz

# Add nerd CLI
RUN curl -L https://github.com/nerdalize/nerd/releases/download/v0.2.3/nerd-0.2.3-linux.tar.gz | tar xz

COPY testinputs /testinputs

COPY run.sh /run.sh
CMD /run.sh
