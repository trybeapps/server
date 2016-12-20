from flask import Flask, g, session, abort, redirect, url_for, render_template, request, escape, send_from_directory, jsonify
from flask_sqlalchemy import SQLAlchemy
from celery import Celery
import os

# Define the WSGI application object
app = Flask(__name__)

# Set the secret key.  keep this really secret:
app.secret_key = 'ff29b42f8d7d5cbefd272eab3eba6ec8'

# Set the postgres config
app.config['SQLALCHEMY_DATABASE_URI'] = 'postgresql://libreread:libreread@localhost/libreread_dev'
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