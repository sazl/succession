from pprint import pprint
from wiki_api import *

if __name__ == '__main__':
    pprint(wiki_api.get_subcategories('Roman emperors'))