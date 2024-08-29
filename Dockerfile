FROM ubuntu

RUN echo "Asia/Shanghai" > /etc/timezone &&\ 
    apt-get update || apt install -y tzdata &&\
    dpkg-reconfigure -f noninteractive tzdata

# # Add openssh and supervisord
RUN apt install -y bash openssh-server supervisor curl wget

# Change root password to 'password'
RUN echo 'root:password' | chpasswd
ADD authorized_keys /root/.ssh/authorized_keys

# SSH login fix. Otherwise user is kicked off after login
RUN sed -ie 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config

# Add key
RUN ssh-keygen -A

# Expose SSH port
EXPOSE 22

RUN mkdir -p /etc/supervisor.d/ &&\
    mkdir /run/sshd

ADD supervisord.ini /etc/supervisor.d/supervisord.ini

EXPOSE 9000
USER root

COPY cde-server /

RUN chmod +x /cde-server

# Run supervisord
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor.d/supervisord.ini"]
