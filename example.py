from pprint import pprint

import time
import json

from wiki_api.category import Category, get_category, get_subcategories, CategoryJSONEncoder


if __name__ == '__main__':
    category = Category(name='Roman emperors')

    for m in category.get_members():
        for p in m.get_pages():
            pass

    pprint(category)

    with open('output.json', 'w') as f:
        f.write(json.dumps(category, cls=CategoryJSONEncoder))