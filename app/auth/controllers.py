from flask import Blueprint, g, session, abort, redirect, url_for, render_template, request, escape, send_from_directory, jsonify
from app.auth.models import User
from app.book.models import Book
from app import db
from functools import wraps
from sqlalchemy import desc
import bcrypt

auth = Blueprint('auth', __name__, template_folder='templates')

@auth.route('/')
def index():
    if 'email' in session:
        user = User.query.filter_by(email=session['email']).first()
        if user:
            books = user.books.order_by(desc(Book.created_on)).all()
            print books
            return render_template('home.html', user = user, books = books, new_books=[])
    return render_template('landing.html')

@auth.route('/signup', methods=['GET', 'POST'])
def signup():
    if request.method == 'POST':
        name = request.form['name']
        email = request.form['email']
        password = request.form['password'].encode('utf-8')

        password_hash = bcrypt.hashpw(password, bcrypt.gensalt())

        user = User(name, email, password_hash)
        db.session.add(user)
        db.session.commit()
        session['email'] = email

        return redirect(url_for('auth.index'))
    return render_template('signup.html')
    return '''
        <form action="" method="post">
            <p><input type=text name=name></p>
            <p><input type=text name=email></p>
            <p><input type=text name=password></p>
            <p><input type=submit value=sign up></p>
        </form>
        '''

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
    return '''
        <form action="" method="post">
            <p><input type=text name=email></p>
            <p><input type=text name=password></p>
            <p><input type=submit value=Login></p>
        </form>
    '''

@auth.route('/signout')
def logout():
    # remove the email from the session if it's there
    session.pop('email', None)
    return redirect(url_for('auth.index'))
