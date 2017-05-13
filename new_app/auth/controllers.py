from flask import Blueprint, request, render_template, redirect, url_for, jsonify

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
    return render_template('signin.html', signup=True)

@auth.route('/')
def index():
    return redirect(url_for('auth.signin'))