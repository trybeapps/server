from flask import Flask, g, session, abort, redirect, url_for, render_template, request, escape, send_from_directory, jsonify
from flask_mail import Mail
from flask_sqlalchemy import SQLAlchemy
from celery import Celery
import os

# Define the WSGI application object
app = Flask(__name__)

# Debug mode on
app.config['DEBUG'] = True

# Set the secret key.  keep this really secret:
app.config['SECRET_KEY'] = 'ff29b42f8d7d5cbefd272eab3eba6ec'

# Set the SECURITY_PASSWORD_SALT for email confirmation
app.config['SECURITY_PASSWORD_SALT'] = 'precious'

# Configure mail SMTP
app.config['MAIL_SERVER'] = 'smtp.zoho.com'
app.config['MAIL_PORT'] = 465
app.config['MAIL_USE_SSL'] = True
app.config['MAIL_USE_TLS'] = False
app.config['MAIL_USERNAME'] = os.environ.get('MAIL_USERNAME')
app.config['MAIL_PASSWORD'] = os.environ.get('MAIL_PASSWORD')
app.config['MAIL_DEFAULT_SENDER'] = os.environ.get('MAIL_USERNAME')

mail = Mail(app)

# Set the postgres config
#app.config['SQLALCHEMY_DATABASE_URI'] = 'postgresql://libreread:libreread@localhost/libreread_dev'
db_path = os.path.join(os.path.dirname(__file__), 'libreread.db')
db_uri = 'sqlite:///{}'.format(db_path)
print db_uri
app.config['SQLALCHEMY_DATABASE_URI'] = db_uri
db = SQLAlchemy(app)

# Set the upload path
APP_ROOT = os.path.dirname(os.path.abspath(__file__))
UPLOAD_FOLDER = os.path.join(APP_ROOT, 'uploads')
app.config['UPLOAD_FOLDER'] = UPLOAD_FOLDER

# Allowed extensions for file upload
app.config['ALLOWED_EXTENSIONS'] = set(['pdf'])

# Celery config
app.config['CELERY_BROKER_URL'] = 'amqp://guest:guest@localhost:5672//'
app.config['CELERY_RESULT_BACKEND'] = 'amqp://guest:guest@localhost:5672//'

celery = Celery(app.name, broker=app.config['CELERY_BROKER_URL'])
celery.conf.update(app.config)

# Import a module / component using its blueprint handler variable (eg: auth)
from app.auth.controllers import auth
from app.book.controllers import book
from app.collection.controllers import collection

# Register bludeprint(s)
app.register_blueprint(auth)
app.register_blueprint(book)
app.register_blueprint(collection)
