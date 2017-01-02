from flask import Blueprint, g, session, abort, redirect, url_for, render_template, request, escape, send_from_directory, jsonify
from app.auth.controllers import login_required
from app.auth.models import User
from app.book.models import Book
from app import app, celery, db
from werkzeug.utils import secure_filename
import requests
import os
import shutil
import json
import base64
import subprocess
from datetime import datetime

book = Blueprint('book', __name__, template_folder='templates')

def allowed_file(filename):
    return '.' in filename and \
           filename.rsplit('.', 1)[1] in app.config['ALLOWED_EXTENSIONS']

@book.route('/book-upload', methods=['GET', 'POST'])
def upload_file():

    import requests
    import json

    payload = json.dumps({
        'description': 'Process documents',
        'processors': [
            {
                'attachment': {
                    'field': 'thedata',
                    'indexed_chars': -1
                }
            },
            {
                'set': {
                    'field': 'attachment.title',
                    'value': '{{ title }}'
                }
            },
            {
                'set': {
                    'field': 'attachment.page',
                    'value': '{{ page }}'
                }
            }
        ]
    })

    print payload

    r = requests.put('http://elasticsearch:9200/_ingest/pipeline/attachment', data=payload)

    print r.text

    if request.method == 'POST':
        args= []
        for i in range(len(request.files)):
          file = request.files['file['+str(i)+']']
          if file.filename == '':
              print ('No selected file')
              return redirect(request.url)
          if file and allowed_file(file.filename):

              # Check if the file already exists then return
              user = User.query.filter_by(email=session['email']).first()
              book = Book.query.filter_by(user_id=user.id, filename=file.filename).first()
              print book
              if book:
                  print 'book already exists'
                  return 'book already exists'

              # Generate a new filename with datetime for uniqueness when other user has the same filename
              filename = secure_filename(file.filename)
              filename_gen = filename.split('.pdf')[0] + '_' + "{:%M%S%s}".format(datetime.now()) + '.pdf'
              print filename_gen
              if not os.path.exists(app.config['UPLOAD_FOLDER']):
                  os.makedirs(os.path.join(app.config['UPLOAD_FOLDER'], 'images/'))
              file_path = os.path.join(app.config['UPLOAD_FOLDER'], filename_gen)
              file.save(file_path)
              print 'file saved successfully'

              info = _pdfinfo(file_path)
              print (info)

              # Check if info['Title'] exists else pass the filename as the title in the db
              try:
                  title = info['Title']
              except KeyError, e:
                  print 'No Title found, so passing the filename as title'
                  title = filename.split('.pdf')[0]

              # Check if info['Author'] exists else pass Unknown as the author in the db
              try:
                  author = info['Author']
              except KeyError, e:
                  print 'No Author found, so passing unknown'
                  author = 'unknown'


              img_folder = 'images/' + '_'.join(title.split(' '))
              cover_path = os.path.join(app.config['UPLOAD_FOLDER'], img_folder)

              _gen_cover(file_path, cover_path)

              url = '/b/' + filename_gen
              cover = '/b/cover/' + '_'.join(title.split(' ')) + '-001-000.png'
              print cover

              book = Book(title=title, filename=file.filename, author=author, url=url, cover=cover, pages=info['Pages'], current_page=0)

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

              # Send the request to ElasticSearch on elasticsearch:9200
              r = requests.put('http://elasticsearch:9200/lr_index/book_info/' + str(book.id), data=book_info)
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
        # _feed_content.delay(args=args)
        _feed_content(args=args)
        return 'success'
    else:
        return redirect(url_for('index'))

# @celery.task()
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
            r = requests.put('http://elasticsearch:9200/lr_index/book_detail/' + str(arg['user_id']) + '_' + str(arg['book']['book_id']) + '_' + str(i) + '?pipeline=attachment', data=book_detail)
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

    cmd = '/usr/bin/pdfinfo'
    if not osp.exists(cmd):
        raise RuntimeError('System command not found: %s' % cmd)

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

@book.route('/uploads/<filename>')
@login_required
def uploaded_file(filename):
    print app.config['UPLOAD_FOLDER']
    return send_from_directory(app.config['UPLOAD_FOLDER'],
                               filename)

@book.route('/b/<filename>')
@login_required
def send_book(filename):
    # return send_from_directory('uploads', filename)
    file_path = '/b/' + filename
    if 'email' in session:
        user = User.query.filter_by(email=session['email']).first()
        if user:
            books = user.books.all()
            for book in books:
                print book.url
                print file_path
                if book.url == file_path:
                    return render_template('viewer.html', book_id=book.id)
    return redirect(url_for('auth.index'))

@book.route('/b/cover/<filename>')
def send_book_cover(filename):
    return send_from_directory('uploads/images', filename)

@book.route('/b/delete/<id>')
@login_required
def delete_book(id):
    book = db.session.query(Book).get(id)

    # Delete the book from elastic search
    r = requests.delete('http://elasticsearch:9200/lr_index/book_info/' + str(book.id))
    print r.text

    # Delete all pages from the elastic search
    user = User.query.filter_by(email=session['email']).first()
    for i in range(1,book.pages+1):
        r = requests.delete('http://elasticsearch:9200/lr_index/book_detail/' + str(user.id) + '_' + str(book.id) + '_' + str(i))
        print r.text
    db.session.delete(book)
    db.session.commit()
    return 'successfully deleted'

@book.route('/autocomplete', methods=['GET'])
@login_required
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

    r = requests.get('http://elasticsearch:9200/lr_index/book_info/_search', data=payload)
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

    r = requests.get('http://elasticsearch:9200/lr_index/book_detail/_search', data=payload)
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
