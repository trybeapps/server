from flask import Flask
app = Flask(__name__)

url_for('static', filename='style.css')

@app.route('/')
def hello_world():
    return 'Hello, World!'
