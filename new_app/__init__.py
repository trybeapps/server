from flask import Flask
from flask_sqlalchemy import SQLAlchemy
import os

# Define the WSGI application object
app = Flask(__name__)

# Debug mode on
app.config['DEBUG'] = True

# Set the secret key.  keep this really secret:
app.config['SECRET_KEY'] = 'ff29b42f8d7d5cbefd272eab3eba6ec'

# Set the SECURITY_PASSWORD_SALT for email confirmation
app.config['SECURITY_PASSWORD_SALT'] = 'precious'

# Set the DB config
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False
db_path = os.path.join(os.path.dirname(__file__), 'libreread.db')
db_uri = 'sqlite:///{}'.format(db_path)
print db_uri
app.config['SQLALCHEMY_DATABASE_URI'] = db_uri
db = SQLAlchemy(app)

# Import a module / component using its blueprint handler variable (eg: auth)
from new_app.auth.controllers import auth

# Register bludeprint(s)
app.register_blueprint(auth)