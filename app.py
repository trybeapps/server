from flask import Flask, g, session, abort, redirect, url_for, render_template, request, escape, send_from_directory, jsonify
from werkzeug.utils import secure_filename
from functools import wraps
from sqlalchemy import desc
from flask_sqlalchemy import SQLAlchemy
from celery import Celery
import hashlib
import requests
import os
import shutil
import json
import base64
import datetime
import subprocess
import bcrypt
from datetime import datetime

APP_ROOT = os.path.dirname(os.path.abspath(__file__))
UPLOAD_FOLDER = os.path.join(APP_ROOT, 'uploads')
ALLOWED_EXTENSIONS = set(['pdf'])

app = Flask(__name__)

# celery config
app.config['CELERY_BROKER_URL'] = 'redis://localhost:6379/0'
app.config['CELERY_RESULT_BACKEND'] = 'redis://localhost:6379/0'

celery = Celery(app.name, broker=app.config['CELERY_BROKER_URL'])
celery.conf.update(app.config)

# set the upload path
app.config['UPLOAD_FOLDER'] = UPLOAD_FOLDER

# set the secret key.  keep this really secret:
app.secret_key = 'ff29b42f8d7d5cbefd272eab3eba6ec8'

app.config['SQLALCHEMY_DATABASE_URI'] = 'postgresql://localhost/libreread_dev'
db = SQLAlchemy(app)

from models import User, Book, Collection

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
        user = User.query.filter_by(email=session['email']).first()
        if user:
            books = user.books.order_by(desc(Book.created_on)).all()
            print books
            return render_template('home.html', user = user, books = books, new_books=[])
    return render_template('landing.html')

@app.route('/signup', methods=['GET', 'POST'])
def signup():
    if 'email' in session:
        return redirect(url_for('index'))
    else:
        if request.method == 'POST':
            name = request.form['name']
            email = request.form['email']
            password = request.form['password']

            password_hash = bcrypt.hashpw(password, bcrypt.gensalt())

            user = User(name, email, password_hash)
            db.session.add(user)
            db.session.commit()
            session['email'] = email

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

def allowed_file(filename):
    return '.' in filename and \
           filename.rsplit('.', 1)[1] in ALLOWED_EXTENSIONS

@app.route('/uploads/<filename>')
def uploaded_file(filename):
    return send_from_directory(app.config['UPLOAD_FOLDER'],
                               filename)

@app.route('/book-upload', methods=['GET', 'POST'])
def upload_file():
    if request.method == 'POST':
        print 'coming'
        args= []
        for i in range(len(request.files)):
          file = request.files['file['+str(i)+']']
          if file.filename == '':
              print ('No selected file')
              return redirect(request.url)
          if file and allowed_file(file.filename):
              filename = secure_filename(file.filename)
              filename = filename.split('.pdf')[0] + '_' + "{:%M%S%s}".format(datetime.now()) + '.pdf'
              print filename
              file_path = os.path.join(app.config['UPLOAD_FOLDER'], filename)
              file.save(file_path)

              info = _pdfinfo(file_path)
              print (info)

              img_folder = 'images/' + '_'.join(info['Title'].split(' '))
              cover_path = os.path.join(app.config['UPLOAD_FOLDER'], img_folder)

              _gen_cover(file_path, cover_path)

              url = '/b/' + filename
              cover = '/b/cover/' + '_'.join(info['Title'].split(' ')) + '-001-000.png'
              print cover

              book = Book(title=info['Title'], author=info['Author'], url=url, cover=cover, pages=info['Pages'], current_page=0)

              user = User.query.filter_by(email=session['email']).first()
              user.books.append(book)
              db.session.add(user)
              db.session.add(book)
              db.session.commit()

              print book.id
              # Feeding pdf content into ElasticSearch
              # Encode the pdf file and add it to the index
              pdf_data = _pdf_encode(file_path)

              # Set the payload in json
              book_info = json.dumps({
                'title': book.title,
                'author': book.author,
                'url': book.url,
                'cover': book.cover
              })

              # Send the request to ElasticSearch on localhost:9200
              r = requests.put('http://localhost:9200/lr_index/book_info/' + str(book.id), data=book_info)
              print r.text

              # Feed content to Elastic as a background job with celery
              args.append({'user_id': user.id,
                'book': {
                    'book_id': book.id,
                    'book_title': book.title,
                    'book_author': book.author,
                    'book_cover': book.cover,
                    'book_url': book.url,
                    'book_pages': book.pages
                },
                'file_path': file_path
              })

              print user.books

              print ('Book uploaded successfully!')
        _feed_content.delay(args=args)
        return 'success'
    else:
        return redirect(url_for('index'))

@celery.task()
def _feed_content(args):

    for arg in args:
        # Make directory for adding the pdf separated files
        directory = os.path.join(app.config['UPLOAD_FOLDER'], 'splitpdf' + '_' + str(arg['user_id']) + '_' + "{:%M%S%s}".format(datetime.now()))
        print directory
        if not os.path.exists(directory):
            os.makedirs(directory)

        _pdf_separate(directory, arg['file_path'])

        for i in range(1,int(arg['book']['book_pages'])+1):
            pdf_data = _pdf_encode(directory+'/'+str(i)+'.pdf')
            book_detail = json.dumps({
                'thedata': pdf_data,
                'title': arg['book']['book_title'],
                'author': arg['book']['book_author'],
                'url': arg['book']['book_url'],
                'cover': arg['book']['book_cover'],
                'page': i,
            })
            # feed data in id = userid_bookid_pageno
            r = requests.put('http://localhost:9200/lr_index/book_detail/' + str(arg['user_id']) + '_' + str(arg['book']['book_id']) + '_' + str(i) + '?pipeline=attachment', data=book_detail)
            print r.text

        # Remove the splitted pdfs as it is useless now
        shutil.rmtree(directory)

def _pdf_separate(directory, file_path):
    subprocess.call('pdfseparate ' + file_path + ' ' + directory + '/%d.pdf', shell=True)

def _pdfinfo(infile):
    """
    Wraps command line utility pdfinfo to extract the PDF meta information.
    Returns metainfo in a dictionary.
    sudo apt-get install poppler-utils
    This function parses the text output that looks like this:
        Title:          PUBLIC MEETING AGENDA
        Author:         Customer Support
        Creator:        Microsoft Word 2010
        Producer:       Microsoft Word 2010
        CreationDate:   Thu Dec 20 14:44:56 2012
        ModDate:        Thu Dec 20 14:44:56 2012
        Tagged:         yes
        Pages:          2
        Encrypted:      no
        Page size:      612 x 792 pts (letter)
        File size:      104739 bytes
        Optimized:      no
        PDF version:    1.5
    """
    import os.path as osp
    import subprocess

    cmd = '/usr/bin/pdfinfo'
    # if not osp.exists(cmd):
    #     raise RuntimeError('System command not found: %s' % cmd)

    if not osp.exists(infile):
        raise RuntimeError('Provided input file not found: %s' % infile)

    def _extract(row):
        """Extracts the right hand value from a : delimited row"""
        return row.split(':', 1)[1].strip()

    output = {}

    labels = ['Title', 'Author', 'Creator', 'Producer', 'CreationDate',
              'ModDate', 'Tagged', 'Pages', 'Encrypted', 'Page size',
              'File size', 'Optimized', 'PDF version']

    cmd_output = subprocess.check_output(['pdfinfo', infile])
    for line in cmd_output.splitlines():
        for label in labels:
            if label in line:
                output[label] = _extract(line)

    return output

def _gen_cover(file_path, cover_path):
    print file_path
    print cover_path
    subprocess.call('pdfimages -p -png -f 1 -l 2 ' + file_path + ' ' + cover_path, shell=True)

def _pdf_encode(pdf_filename):
    return base64.b64encode(open(pdf_filename,"rb").read());

@app.route('/b/<filename>')
def send_book(filename):
    # return send_from_directory('uploads', filename)
    file_path = '/b/' + filename
    if 'email' in session:
        user = User.query.filter_by(email=session['email']).first()
        if user:
            books = user.books.all()
            for book in books:
                if book.url == file_path:
                    return render_template('viewer.html')
    return redirect(url_for('index'))

@app.route('/b/cover/<filename>')
def send_book_cover(filename):
    return send_from_directory('uploads/images', filename)

@app.route('/autocomplete', methods=['GET'])
def search_books():
    query = request.args.get('term')
    print query

    suggestions = []

    payload = json.dumps({
        '_source': ['title', 'author', 'url', 'cover'],
        'query': {
            'multi_match': {
                'query': query,
                'fields': ['title', 'author']
            }
        }
    })

    r = requests.get('http://localhost:9200/lr_index/book_info/_search', data=payload)
    data = json.loads(r.text)
    print (data)

    hits = data['hits']['hits']
    total = int(data['hits']['total'])

    metadata = []

    for hit in hits:
        title = hit['_source']['title']
        author = hit['_source']['author']
        url = hit['_source']['url']
        cover = hit['_source']['cover']

        metadata.append({
            'title': title, 'author': author, 'url': url, 'cover': cover
        })


    suggestions.append(metadata)

    payload = json.dumps({
        '_source': ['title', 'author', 'url', 'cover', 'page'],
        'query': {
            'match_phrase': {
                'attachment.content': query
            }
        },
        'highlight': {
            'fields': {
                'attachment.content': {
                    'fragment_size': 150,
                    'number_of_fragments': 3,
                    'no_match_size': 150
                }
            }
        }
    })

    r = requests.get('http://localhost:9200/lr_index/book_detail/_search', data=payload)
    data = json.loads(r.text)
    print (data)

    hits = data['hits']['hits']
    total = int(data['hits']['total'])

    content = []

    for hit in hits:
        title = hit['_source']['title']
        author = hit['_source']['author']
        url = hit['_source']['url']
        cover = hit['_source']['cover']
        page = hit['_source']['page']
        data = hit['highlight']['attachment.content']
        if len(data):
            for i in data:
                content.append({
                    'title': title, 'author': author, 'url': url, 'cover': cover, 'page': page, 'data': i
                })

    suggestions.append(content)
    print suggestions

    return jsonify(suggestions)

@app.route('/collections', methods=['GET', 'POST'])
def collections():
    if request.method == 'GET':
        collections = Collection.query.order_by(desc(Collection.id)).all()
        print collections
        return render_template('collection.html', current_page='collections', collections=collections)

@app.route('/collections/new', methods=['GET', 'POST'])
def new_collection():
    if 'email' in session:
        if request.method == 'GET':
            user = User.query.filter_by(email=session['email']).first()
            books = user.books.order_by(desc(Book.created_on)).all()
            print books
            return render_template('new_collection.html', current_page='collections', books=books)
        else:
            title = request.form.get('title', None)
            checked_list = request.form.getlist('book')
            collection = Collection(title=title)
            for i in checked_list:
                book = Book.query.filter_by(id=i).first()
                collection.books.append(book)
                print collection
                db.session.add(collection)
                db.session.add(book)
                db.session.commit()
            return 'success'
    return redirect(url_for('login'))
