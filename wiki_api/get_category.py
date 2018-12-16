#!/usr/bin/python3

"""
	get_category_items.py

    MediaWiki Action API Code Samples
    Demo of `Categorymembers` module : List twenty items in a category.
    MIT license
"""

import requests

def get_category(category, limit=10):
    S = requests.Session()

    URL = "https://en.wikipedia.org/w/api.php"

    PARAMS = {
        'action': "query",
        'list': "categorymembers",
        'cmtitle': f"Category:{category}",
        'cmlimit': limit,
        'format': "json"
    }

    R = S.get(url=URL, params=PARAMS)
    DATA = R.json()
    return DATA
