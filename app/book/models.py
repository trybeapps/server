from app import db

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