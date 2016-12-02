from flask import Flask, g, session, abort, redirect, url_for, render_template, request, escape
from functools import wraps
from flask_sqlalchemy import SQLAlchemy
import hashlib
import os
import bcrypt

app = Flask(__name__)
app.config['SQLALCHEMY_DATABASE_URI'] = 'postgresql://localhost/libreread_dev'
db = SQLAlchemy(app)

# Create our database model
class User(db.Model):
    __tablename__ = "users"
    id = db.Column(db.Integer, primary_key=True)
    name = db.Column(db.String(200))
    email = db.Column(db.String(120), unique=True)
    password_hash = db.Column(db.String(80))

    def __init__(self, name, email, password_hash):
        self.name = name
        self.email = email
        self.password_hash = password_hash

    def __repr__(self):
        return '<Email %r>' % self.email

@app.before_request
def before_request():
    if 'email' in session:
        g.user = session['email']
    else:
        g.user = None

def login_required(f):
    @wraps(f)
    def decorated_function(*args, **kwargs):
        if g.user is None:
            return redirect(url_for('login', next=request.url))
        return f(*args, **kwargs)
    return decorated_function

@app.route('/')
def index():
    if 'email' in session:
        print (session)
        return render_template('home.html')
    return render_template('landing.html')

@app.route('/signup', methods=['GET', 'POST'])
def signup():
    if request.method == 'POST':
        name = request.form['name']
        email = request.form['email']
        password = request.form['password']

        password_hash = bcrypt.hashpw(password, bcrypt.gensalt())

        user = User(name, email, password_hash)
        db.session.add(user)
        db.session.commit()
        session['email'] = email
        users = User.query.all()

        print (users)

        return redirect(url_for('index'))
    return '''
        <form action="" method="post">
            <p><input type=text name=name></p>
            <p><input type=text name=email></p>
            <p><input type=text name=password></p>
            <p><input type=submit value=sign up></p>
        </form>
    '''

@app.route('/signin', methods=['GET', 'POST'])
def login():
    if request.method == 'POST':
        email = request.form['email']
        password = request.form['password']

        user = User.query.filter_by(email=email).first()

        if user is not None:
            if bcrypt.hashpw(password, user.password_hash) == user.password_hash:
                session['email'] = email
                return redirect(url_for('index'))
    return '''
        <form action="" method="post">
            <p><input type=text name=email></p>
            <p><input type=text name=password></p>
            <p><input type=submit value=Login></p>
        </form>
    '''

@app.route('/signout')
def logout():
    # remove the email from the session if it's there
    session.pop('email', None)
    return redirect(url_for('index'))

# set the secret key.  keep this really secret:
app.secret_key = 'ff29b42f8d7d5cbefd272eab3eba6ec8'
