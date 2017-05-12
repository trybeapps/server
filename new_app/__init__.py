from flask import Flask

# Define the WSGI application object
app = Flask(__name__)

# Debug mode on
app.config['DEBUG'] = True

# Set the secret key.  keep this really secret:
app.config['SECRET_KEY'] = 'ff29b42f8d7d5cbefd272eab3eba6ec'

# Set the SECURITY_PASSWORD_SALT for email confirmation
app.config['SECURITY_PASSWORD_SALT'] = 'precious'

# Import a module / component using its blueprint handler variable (eg: auth)
from new_app.auth.controllers import auth

# Register bludeprint(s)
app.register_blueprint(auth)