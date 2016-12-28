############################################################
# Dockerfile to build LibreRead
# Based on Debian
############################################################

# Set the base image to Debian
FROM debian

# File Author / Maintainer
MAINTAINER LibreRead Nirmal

# Update the repository sources list
RUN apt-get update

# Install basic applications
RUN apt-get install -y tar git curl nano wget dialog net-tools build-essential

# Install Python and Basic Python Tools
RUN apt-get install -y python python-dev python-distribute python-pip

# Install dependencies for psycopg2 and py-bcrypt
RUN apt-get install -y libssl-dev libffi-dev

# Clone the repository from github
RUN git clone git@github.com:mysticmode/LibreRead.git root/LibreRead

# Set the default directory where CMD will execute
WORKDIR root/LibreRead

# Get pip to download and install requirements:
RUN pip install -r requirements.txt

# Expose ports
EXPOSE 8000

# Set the command to execute python server
CMD python wsgi.py