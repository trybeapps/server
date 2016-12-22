from flask import Blueprint, g, session, abort, redirect, url_for, render_template, request, escape, send_from_directory, jsonify
from sqlalchemy import desc
from app.auth.models import User
from app.book.models import Book
from app.collection.models import Collection
from app import app, celery, db
from werkzeug.utils import secure_filename
import requests
import os
import shutil
import json
import base64
import subprocess
from datetime import datetime

collection = Blueprint('collection', __name__, template_folder='templates')

@collection.route('/collections', methods=['GET', 'POST'])
def collections():
    if request.method == 'GET':
        collections = Collection.query.order_by(desc(Collection.id)).all()
        print collections
        user = User.query.filter_by(email=session['email']).first()
        return render_template('collection.html', user=user, current_page='collections', collections=collections)

@collection.route('/collections/new', methods=['GET', 'POST'])
def new_collection():
    if 'email' in session:
        if request.method == 'GET':
            user = User.query.filter_by(email=session['email']).first()
            books = user.books.order_by(desc(Book.created_on)).all()
            print books
            return render_template('new_collection.html', user=user, current_page='collections', books=books)
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

@collection.route('/collections/<id>', methods=['GET'])
def collection_detail(id):
    if 'email' in session:
        collection = Collection.query.filter_by(id=int(id)).first()
        user = User.query.filter_by(email=session['email']).first()
        return render_template('collection_detail.html', user=user, collection=collection)

@collection.route('/templates/<path:path>')
def send_js(path):
    return send_from_directory('templates', path)
