from typing import List
from dataclasses import dataclass, field

import copy

from wiki_api import base

@dataclass
class Category:
    name: str = None
    namespace: int = 0
    pageid: int = 0
    title: str = None
    members: List['Category'] = field(default_factory=list)

    @property
    def cmtitle(self):
        if self.title is None:
            return f'Category:{self.name}'
        return self.title

    @classmethod
    def from_dict(cls, data):
        return cls(title=data['title'], namespace=data['ns'], pageid=data['pageid'])



def get_subcategories(category, limit=10):
    params = {
        'action': "query",
        'list': "categorymembers",
        'cmtitle': category.cmtitle,
        'cmlimit': limit,
        'cmtype': "subcat",
        'format': "json"
    }

    category_members = base.query(params)
    result = copy.deepcopy(category)

    for category_member in category_members:
        data = category_member['categorymembers']
        for category in data:
            cat = Category.from_dict(category)
            result.members.append(cat)

    return result


def get_category(category, limit=10):
    params = {
        'action': "query",
        'list': "categorymembers",
        'cmtitle': category.cmtitle,
        'cmlimit': limit,
        'format': "json"
    }

    return base.query(params)