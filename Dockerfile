############################################################
# Dockerfile to build LibreRead
# Based on Debian
############################################################

# Set the base image to Debian
FROM debian:8

# File Author / Maintainer
MAINTAINER LibreRead Nirmal

# Update the repository sources list
RUN apt-get update

# Install basic applications
RUN apt-get install -y tar git curl nano wget dialog net-tools build-essential

# Install Python and Basic Python Tools
RUN apt-get install -y python python-dev python-distribute python-pip poppler-utils

# Install dependencies for py-bcrypt
RUN apt-get install -y libssl-dev libffi-dev

# Clone the repository from github
RUN git clone https://github.com/mysticmode/LibreRead.git opt/LibreRead

# Set the default directory where CMD will execute
WORKDIR opt/LibreRead

# Get pip to download and install requirements:
RUN pip install -r requirements.txt

# Create db file
RUN touch app/libreread.db

# File permission for db file
RUN chmod 777 app/libreread.db

# Set the environment variables for smtp mail server
ENV MAIL_USERNAME email_address
ENV MAIL_PASSWORD password
ENV MAIL_DEFAULT_SENDER email_address

# Set the command to create db
RUN python db_create.py

# Init ES attachment
RUN python config/elastic/init_attachment.py

# Init ES Index
RUN python config/elastic/init_index.py

# Expose ports
EXPOSE 8000

# Set the command to execute python server
CMD python wsgi.py
