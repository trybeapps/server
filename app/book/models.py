from app import db

class Book(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    title = db.Column(db.String(200))
    filename = db.Column(db.String(200))
    author = db.Column(db.String(200))
    url = db.Column(db.String(200))
    cover = db.Column(db.String(200))
    pages = db.Column(db.Integer)
    current_page = db.Column(db.Integer)
    created_on = db.Column(db.DateTime, server_default=db.func.now())
    public = db.Column(db.Boolean, default=True)
    user_id = db.Column(db.Integer, db.ForeignKey('user.id'))
    highlights = db.relationship('Highlight', backref='book', lazy='dynamic')

    def __init__(self, title, filename, author, url, cover, pages, current_page):
        self.title = title
        self.filename = filename
        self.author = author
        self.url = url
        self.cover = cover
        self.pages = pages
        self.current_page = current_page

    def __repr__(self):
        return '<Book %r>' % (self.title)

class Highlight(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    page_container = db.Column(db.String(200))
    nth_child = db.Column(db.Integer)
    html = db.Column(db.String(1200))
    book_id = db.Column(db.Integer, db.ForeignKey('book.id'))

    def __init__(self, page_container, nth_child, html):
        self.page_container = page_container
        self.nth_child = nth_child
        self.html = html

    def __repr__(self):
        return '<Highlight %r>' % (self.html)
