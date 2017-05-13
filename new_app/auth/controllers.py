from flask import Blueprint, request, render_template, redirect, url_for, jsonify
import os

auth = Blueprint('auth', __name__, template_folder='templates')


@auth.route('/signin', methods=['GET', 'POST'])
def signin():
    if request.method == 'POST':
        email = request.form['email']
        password = request.form['password'].encode('utf-8')

        return jsonify(
    	    email=email
        )
    return render_template('signin.html', signup=False)

@auth.route('/signup', methods=['GET', 'POST'])
def signup():
    if request.method == 'POST':
        name = request.form['name']
        email = request.form['email']
        password = request.form['password'].encode('utf-8')

        return jsonify(
            email=email
        )
    return render_template('signin.html', signup=True)

@auth.route('/')
def index():
    return render_template('home.html', path='index')

@auth.route('/collections')
def collections():
	return render_template('home.html', path='collections')