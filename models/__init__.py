from app import db

collections = db.Table('collections',
    db.Column('collection_id', db.Integer, db.ForeignKey('collection.id')),
    db.Column('book_id', db.Integer, db.ForeignKey('book.id'))
)

# Create our database model
class User(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    name = db.Column(db.String(200))
    email = db.Column(db.String(120), unique=True)
    password_hash = db.Column(db.String(80))
    books = db.relationship('Book', backref='user', lazy='dynamic')

    def __init__(self, name, email, password_hash):
        self.name = name
        self.email = email
        self.password_hash = password_hash

    def __repr__(self):
        return '<Email %r>' % self.email

class Book(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    title = db.Column(db.String(200))
    author = db.Column(db.String(200))
    url = db.Column(db.String(200))
    cover = db.Column(db.String(200))
    pages = db.Column(db.Integer)
    current_page = db.Column(db.Integer)
    created_on = db.Column(db.DateTime, server_default=db.func.now())
    public = db.Column(db.Boolean, default=True)
    user_id = db.Column(db.Integer, db.ForeignKey('user.id'))

    def __init__(self, title, author, url, cover, pages, current_page):
        self.title = title
        self.author = author
        self.url = url
        self.cover = cover
        self.pages = pages
        self.current_page = current_page

    def __repr__(self):
        return '<Book %r>' % (self.title)

class Collection(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    title = db.Column(db.String(200))
    books = db.relationship('Book', secondary=collections, backref=db.backref('collections', lazy='dynamic'))

    def __init(self, title):
        self.title = title

    def __repr__(self):
        return '<Collection %r>' % (self.title)
