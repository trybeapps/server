# LibreRead
Browser-based Ebook Reader with Full-text search through all the ebooks

###Setup
  - Clone this repo and enable virtualenv. Then run <br/>
    `venv/bin/pip install -r requirements.txt` <br/>
  
  - Setup and start the Postgres server <br/>
  
  - Go to python shell and initiate the db <br/>
    `from app import db` <br/>
    `db.create_all()` <br/>

  - Download and run Elastic Search <br/>
    `./bin/elasticsearch` <br/>
    
  - Ingest attachment in Elastic Search <br/>
    `python config/elastic/init_attachment.py` <br/>
    `python config/elastic/init_index.py` <br/>
  
  - Start the app <br/>
    `flask run`
