from flask import Blueprint, g, session, abort, redirect, url_for, render_template, request, escape, send_from_directory, jsonify
from app.auth.models import User, Invite
from app.book.models import Book
from app import app
from app import db
from app import mail
from functools import wraps
from flask_mail import Message
from functools import wraps
from sqlalchemy import desc
import bcrypt
from itsdangerous import URLSafeTimedSerializer

auth = Blueprint('auth', __name__, template_folder='templates')
def login_required(f):
    @wraps(f)
    def decorated_function(*args, **kwargs):
        user = User.query.filter_by(email=session['email']).first()
        if 'email' not in session:
            return redirect(url_for('auth.login'))
        elif user:
            if not user.confirmed:
                return redirect(url_for('auth.unconfirmed'))
        return f(*args, **kwargs)
    return decorated_function

def generate_confirmation_token(email):
    serializer = URLSafeTimedSerializer(app.config['SECRET_KEY'])
    return serializer.dumps(email, salt=app.config['SECURITY_PASSWORD_SALT'])

def confirm_token(token, expiration=3600):
    serializer = URLSafeTimedSerializer(app.config['SECRET_KEY'])
    try:
        email = serializer.loads(
            token,
            salt=app.config['SECURITY_PASSWORD_SALT'],
            max_age=expiration
        )
    except:
        return False
    return email

@auth.route('/confirm/<token>')
def confirm_email(token):
    try:
        email = confirm_token(token)
    except:
        return 'The confirmation link is invalid or has expired.'
    user = User.query.filter_by(email=email).first_or_404()
    if user.confirmed:
        return 'Account already confirmed. Please login.'
    else:
        user.confirmed = True
        session['email_confirmed'] = True
        db.session.add(user)
        db.session.commit()
        return 'You have confirmed your account. Thanks!'

def send_email(to, subject, template):
    msg = Message(
        subject,
        recipients=[to],
        html=template,
        sender=app.config['MAIL_DEFAULT_SENDER']
    )
    mail.send(msg)

@auth.route('/unconfirmed')
def unconfirmed():
    if 'email' not in session:
        return redirect(url_for('auth.login'))
    elif 'email_confirmed' in session:
        return redirect(url_for('auth.index'))
    print('Please confirm your account!')
    return render_template('unconfirmed.html')

@auth.route('/resend-confirmation')
def resend_confirmation():
    if 'email' in session:
        token = generate_confirmation_token(session['email'])
        confirm_url = url_for('auth.confirm_email', token=token, _external=True)
        html = render_template('activate.html', confirm_url=confirm_url)
        subject = 'Please confirm your email'
        send_email(session['email'], subject, html)
        print ('A new confirmation email has been sent.')
    return redirect(url_for('auth.unconfirmed'))

@auth.route('/')
@login_required
def index():
    user = User.query.filter_by(email=session['email']).first()
    if user:
        books = user.books.order_by(desc(Book.created_on)).all()
        print books
        return render_template('home.html', user = user, books = books, new_books=[])
    return render_template('landing.html')

@auth.route('/signup', methods=['GET', 'POST'])
def signup():
    if request.method == 'GET':
        user = User.query.all()
        if user:
            return 'Sorry, you have no invitation'
        return render_template('signup.html')
    elif request.method == 'POST':
        name = request.form['name']
        email = request.form['email']
        password = request.form['password'].encode('utf-8')

        password_hash = bcrypt.hashpw(password, bcrypt.gensalt())

        user = User(name, email, password_hash, False)
        db.session.add(user)
        db.session.commit()
        session['email'] = email
        token = generate_confirmation_token(email)
        confirm_url = url_for('auth.confirm_email', token=token, _external=True)
        html = render_template('activate.html', confirm_url=confirm_url)
        subject = "Please confirm your email"
        send_email(user.email, subject, html)

        return redirect(url_for('auth.index'))

@auth.route('/signin', methods=['GET', 'POST'])
def login():
    if request.method == 'POST':
        email = request.form['email']
        password = request.form['password'].encode('utf-8')

        user = User.query.filter_by(email=email).first()

        if user is not None:
            if bcrypt.hashpw(password, user.password_hash.encode('utf-8')) == user.password_hash:
                session['email'] = email
                return redirect(url_for('auth.index'))
    return render_template('signin.html')

@auth.route('/signout')
def logout():
    # remove the email from the session if it's there
    session.pop('email', None)
    return redirect(url_for('auth.index'))

@auth.route('/invite', methods=['GET', 'POST'])
@login_required
def invite():
    user = User.query.filter_by(email=session['email']).first()
    if request.method == 'POST':
        users = request.form['users']
        if users:
            users = users.split(',')
            for email in users:
                invite = Invite.query.filter_by(email=email).first()
                if not invite:
                    invite = Invite(email=email)
                    db.session.add(invite)
                    db.session.commit()

                confirm_url = url_for('auth.invitation', email=invite.email, _external=True)
                html = render_template('invitation.html', confirm_url=confirm_url)
                subject = 'Invitation from LibreRead'
                send_email(invite.email, subject, html)
            return 'Sent invitation successfully!'
    return render_template('invite.html', user=user)

@auth.route('/invitation/<email>', methods=['GET'])
def invitation(email):
    email = Invite.query.filter_by(email=email).first()
    if email:
        return render_template('signup.html')
    else:
        return 'Sorry, You have no invitation'
