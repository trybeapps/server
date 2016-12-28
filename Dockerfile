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

# Docker add init.sql for create user, password and database
FROM library/postgres
ENV POSTGRES_USER libreread
ENV POSTGRES_PASSWORD libreread
ENV POSTGRES_DB libreread_dev

# Update the repository sources list
RUN apt-get update

# Install basic applications
RUN apt-get install -y tar git curl nano wget dialog net-tools build-essential

# Install Python and Basic Python Tools
RUN apt-get install -y python python-dev python-distribute python-pip

# Install Postgres
# RUN apt-get install -y postgresql-9.4 postgresql-client-9.4

# Install dependencies for psycopg2 and py-bcrypt
RUN apt-get install -y libpq-dev libssl-dev libffi-dev

# Clone the repository from github
RUN git clone https://github.com/mysticmode/LibreRead.git home/LibreRead

# Set the default directory where CMD will execute
WORKDIR home/LibreRead

# Get pip to download and install requirements:
RUN pip install -r requirements.txt

# Expose ports
EXPOSE 8000

# Set the command to execute python server
CMD python wsgi.py