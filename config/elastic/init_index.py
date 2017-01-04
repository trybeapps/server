import requests
import json

payload = json.dumps({
    'settings': {
        'number_of_shards': 4,
        'number_of_replicas': 0
    }
})

print payload

r = requests.put('http://localhost:9200/lr_index', data=payload)

print r.text
