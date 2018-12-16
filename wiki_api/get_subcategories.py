#!/usr/bin/python3

"""
    get_subcategories.py

    MediaWiki Action API Code Samples
    Demo of `Categorymembers` module : Get ten subcategories of a category.
    MIT license
"""

import requests

def get_sub_categories(category, limit=20):
    S = requests.Session()

    URL = "https://en.wikipedia.org/w/api.php"

    PARAMS = {
        'action': "query",
        'list': "categorymembers",
        'cmtitle': f"Category:{category}",
        'cmtype': "subcat",
        'cmlimit': limit,
        'format': "json"
    }

    R = S.get(url=URL, params=PARAMS)
    DATA = R.json()
    return DATA
