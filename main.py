from pprint import pprint

import time

from wiki_api.category import Category, get_subcategories


if __name__ == '__main__':
    category = Category(name='Roman emperors')
    result_category = get_subcategories(category)
    for d in result_category.members:
        pprint(d)