from flask import Blueprint, g, session, abort, redirect, url_for, render_template, request, escape, send_from_directory, jsonify
from sqlalchemy import desc
from app.auth.controllers import login_required
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
@login_required
def collections():
    if request.method == 'GET':
        collections = Collection.query.order_by(desc(Collection.id)).all()
        print collections
        user = User.query.filter_by(email=session['email']).first()
        return render_template('collection.html', user=user, current_page='collections', collections=collections)

@collection.route('/collections/new', methods=['GET', 'POST'])
@login_required
def new_collection():
    if request.method == 'GET':
        user = User.query.filter_by(email=session['email']).first()
        books = user.books.order_by(desc(Book.created_on)).all()
        print books
        return render_template('new_collection.html', user=user, current_page='collections', collection=None, edit_collection=False, books=books)
    else:
        title = request.form.get('title', None)
        checked_list = request.form.getlist('book')
        collection = Collection(title=title)
        db.session.commit()
        for i in checked_list:
            book = Book.query.filter_by(id=i).first()
            collection.books.append(book)
            print collection
            db.session.add(collection)
            db.session.add(book)
            db.session.commit()
        return 'success'

@collection.route('/collections/edit/<id>', methods=['GET', 'POST'])
@login_required
def edit_collection(id):
    if request.method == 'GET':
        user = User.query.filter_by(email=session['email']).first()
        collection = Collection.query.filter_by(id=id).first()
        c_books = collection.books
        print c_books
        books = user.books.order_by(desc(Book.created_on)).all()
        print books
        return render_template('new_collection.html', user=user, current_page='collections', collection=collection, edit_collection=True, books=books, c_books=c_books)
    else:
        title = request.form.get('title', None)
        checked_list = request.form.getlist('book')
        collection = db.session.query(Collection).get(id)
        collection.title = title
        collection.books = []
        db.session.commit()
        for i in checked_list:
            book = Book.query.filter_by(id=i).first()
            collection.books.append(book)
            print collection
            db.session.add(collection)
            db.session.add(book)
            db.session.commit()
        return redirect(url_for('collection.collections'))

@collection.route('/collections/delete/<id>')
@login_required
def delete_collection(id):
    collection = db.session.query(Collection).get(id)
    db.session.delete(collection)
    db.session.commit()
    return redirect(url_for('collection.collections'))

@collection.route('/collections/<id>', methods=['GET'])
@login_required
def collection_detail(id):
    collection = Collection.query.filter_by(id=int(id)).first()
    user = User.query.filter_by(email=session['email']).first()
    return render_template('collection_detail.html', user=user, collection=collection)

@collection.route('/templates/<path:path>')
def send_js(path):
    return send_from_directory('templates', path)
