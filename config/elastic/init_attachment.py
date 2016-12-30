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
