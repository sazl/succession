from typing import List
import dataclasses

import json
import copy

from wiki_api import base

def get_pages(category, limit=50):
    params = {
        'action': "query",
        'list': "categorymembers",
        'cmtitle': category.cmtitle,
        'cmlimit': limit,
        'format': "json"
    }

    return list(base.query(params))[0]['categorymembers']

def get_subcategories(category, limit=50):
    params = {
        'action': "query",
        'list': "categorymembers",
        'cmtitle': category.cmtitle,
        'cmlimit': limit,
        'cmtype': "subcat",
        'format': "json"
    }

    return list(base.query(params))[0]['categorymembers']

def get_category(category, limit=50):
    params = {
        'action': "query",
        'list': "categorymembers",
        'cmtitle': category.cmtitle,
        'cmlimit': limit,
        'format': "json"
    }

    return list(base.query(params))


@dataclasses.dataclass
class Page:
    title: str = None
    pageid: int = None


@dataclasses.dataclass
class Category:
    name: str = None
    namespace: int = None
    pageid: int = None
    title: str = None

    members: List['Category'] = dataclasses.field(default_factory=list)
    pages: List[Page] = dataclasses.field(default_factory=list)

    @property
    def cmtitle(self):
        if self.title is None:
            return f'Category:{self.name}'
        return self.title

    @classmethod
    def from_dict(cls, data):
        return cls(title=data['title'], namespace=data['ns'], pageid=data['pageid'])


class CategoryJSONEncoder(json.JSONEncoder):

    def default(self, o):
        if dataclasses.is_dataclass(o):
            return dataclasses.asdict(o)
        return super().default(o)