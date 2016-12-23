from app import db

collections = db.Table('collections',
    db.Column('collection_id', db.Integer, db.ForeignKey('collection.id')),
    db.Column('book_id', db.Integer, db.ForeignKey('book.id'))
)

class Collection(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    title = db.Column(db.String(200))
    books = db.relationship('Book', secondary=collections, lazy='dynamic')

    def __init__(self, title):
        self.title = title

    def __repr__(self):
        return '<Collection %r>' % (self.title)
