from app import db

# Create our database model
class User(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    name = db.Column(db.String(200), nullable=False)
    email = db.Column(db.String(120), unique=True, nullable=False)
    password_hash = db.Column(db.String(80), nullable=False)
    confirmed = db.Column(db.Boolean, nullable=False, default=False)
    # books = db.relationship('Book', backref='user', lazy='dynamic')

    def __init__(self, name, email, password_hash, confirmed):
        self.name = name
        self.email = email
        self.password_hash = password_hash
        self.confirmed = confirmed

    def __repr__(self):
        return '<Email %r>' % self.email