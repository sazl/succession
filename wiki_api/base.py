import requests
import time

BASE_URL = "https://en.wikipedia.org/w/api.php"

def get(params):
    resp = requests.get(url=BASE_URL, params=params)
    data = resp.json()
    return data

def query(request, delay=0.5):
    request['action'] = 'query'
    request['format'] = 'json'
    lastContinue = {}

    while True:
        req = request.copy()
        req.update(lastContinue)

        if delay:
            time.sleep(delay)

        result = get(params=req)
        if 'error' in result:
            raise Error(result['error'])
        if 'warnings' in result:
            print(result['warnings'])
        if 'query' in result:
            yield result['query']
        if 'continue' not in result:
            break
        lastContinue = result['continue']